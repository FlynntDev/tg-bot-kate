package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/FlynntDev/tg-bot-kate/internal/repository"
)

type Service struct {
	repo            *repository.Repository
	BotToken        string
	ChannelUsername string
}

func NewService(repo *repository.Repository, botToken string, channelUsername string) *Service {
	return &Service{repo: repo, BotToken: botToken, ChannelUsername: channelUsername}
}

func (s *Service) ValidateKeyword(keyword string) (string, bool) {
	filePath, err := s.repo.GetFilePathByKeyword(keyword)
	if err != nil {
		return "", false
	}
	return filePath, true
}

func (s *Service) KeywordExists(keyword string) (bool, string) {
	return s.repo.KeywordExists(keyword)
}

func (s *Service) CheckAdmin(userID int) bool {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getChatMember?chat_id=%s&user_id=%d", s.BotToken, s.ChannelUsername, userID)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Ошибка запроса к API: %v", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения ответа: %v", err)
		return false
	}

	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			Status string `json:"status"`
		} `json:"result"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		return false
	}

	log.Printf("Статус пользователя в канале: %s", result.Result.Status)
	if result.Result.Status == "administrator" || result.Result.Status == "creator" {
		_ = s.SetAdmin(userID, true)
		return true
	}

	return false
}

func (s *Service) SetAdmin(userID int, admin bool) error {
	return s.repo.SetAdmin(userID, admin)
}

func (s *Service) GetStatistics() (map[string]int, int) {
	return map[string]int{}, 0
}

func (s *Service) SaveContact(userID int, contact string) error {
	return s.repo.SaveContact(userID, contact)
}

func (s *Service) CheckSubscription(userID int) (bool, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getChatMember?chat_id=%s&user_id=%d", s.BotToken, s.ChannelUsername, userID)
	log.Printf("Проверка подписки: URL = %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Ошибка запроса к API: %v", err)
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения ответа: %v", err)
		return false, err
	}

	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			Status string `json:"status"`
		} `json:"result"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		return false, err
	}

	log.Printf("Статус подписки пользователя: %s", result.Result.Status)
	return result.Result.Status == "member" || result.Result.Status == "administrator" || result.Result.Status == "creator", nil
}

func (s *Service) UpdateSubscription(userID int, subscribed bool) error {
	return s.repo.UpdateSubscription(userID, subscribed)
}

func (s *Service) AddKeywordFile(keyword string, filePath string) error {
	exists, path := s.repo.KeywordExists(keyword)
	if exists {
		return fmt.Errorf("Слово уже существует: %s", path)
	}
	return s.repo.AddKeywordFile(keyword, filePath)
}
