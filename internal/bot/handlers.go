package bot

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	if msg == nil || msg.From == nil {
		return
	}

	switch msg.Text {
	case "/start":
		b.handleStart(msg)
	case "Я подписался!":
		b.handleCheckSubscription(msg)
	case "Статистика":
		if b.svc.CheckAdmin(int(msg.From.ID)) {
			b.handleStatistics(msg)
		} else {
			b.sendMessage(msg.Chat.ID, "У вас нет прав для выполнения этой команды.")
		}
	case "Добавить кодовое слово к файлу":
		if b.svc.CheckAdmin(int(msg.From.ID)) {
			b.sendMessage(msg.Chat.ID, "Пожалуйста, отправьте кодовое слово и путь к файлу в формате: слово путь к файлу")
		} else {
			b.sendMessage(msg.Chat.ID, "У вас нет прав для выполнения этой команды.")
		}
	case "Инфо":
		b.handleInfo(msg)
	default:
		if msg.Contact != nil {
			b.handleContact(msg)
		} else if strings.Contains(msg.Text, " ") && b.svc.CheckAdmin(int(msg.From.ID)) {
			b.handleAddFile(msg)
		} else {
			b.handleKeyword(msg)
		}
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	baseKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("Поделиться контактом"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Инфо"),
		),
	)

	adminKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("Поделиться контактом"),
			tgbotapi.NewKeyboardButton("Инфо"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Статистика"),
			tgbotapi.NewKeyboardButton("Добавить кодовое слово к файлу"),
		),
	)

	isAdmin := b.svc.CheckAdmin(int(msg.From.ID))
	if isAdmin {
		b.sendMessageWithKeyboard(msg.Chat.ID, "Привет! Уровень доступа: Администратор", adminKeyboard)
	} else {
		b.sendMessageWithKeyboard(msg.Chat.ID, "Привет! Поделитесь своим контактом или нажмите 'Инфо'.", baseKeyboard)
	}
}

func (b *Bot) handleContact(msg *tgbotapi.Message) {
	err := b.svc.SaveContact(int(msg.From.ID), msg.Contact.PhoneNumber)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Ошибка сохранения контакта.")
		return
	}

	subscribed, err := b.svc.CheckSubscription(int(msg.From.ID))
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Ошибка проверки подписки.")
		return
	}

	if subscribed {
		b.sendMessage(msg.Chat.ID, "Вы уже подписаны на канал. Введите кодовое слово.")
	} else {
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Я подписался!"),
			),
		)
		b.sendMessageWithKeyboard(msg.Chat.ID, fmt.Sprintf("Спасибо! Подпишитесь на канал: https://t.me/%s и нажмите кнопку 'Я подписался!'", b.svc.ChannelUsername), keyboard)
	}
}

func (b *Bot) handleCheckSubscription(msg *tgbotapi.Message) {
	subscribed, err := b.svc.CheckSubscription(int(msg.From.ID))
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Ошибка проверки подписки.")
		return
	}

	if subscribed {
		b.sendMessage(msg.Chat.ID, "Подписка подтверждена! Введите кодовое слово.")
	} else {
		b.sendMessage(msg.Chat.ID, fmt.Sprintf("Пожалуйста, подпишитесь на канал: https://t.me/%s и нажмите кнопку 'Я подписался!'", b.svc.ChannelUsername))
	}
}

func (b *Bot) handleKeyword(msg *tgbotapi.Message) {
	filePath, valid := b.svc.ValidateKeyword(msg.Text)
	if valid {
		err := b.svc.IncrementKeywordCount(msg.Text)
		if err != nil {
			b.sendMessage(msg.Chat.ID, "Ошибка увеличения счетчика использования ключевого слова.")
			return
		}
		b.sendFile(msg.Chat.ID, filePath)
	} else {
		b.sendMessage(msg.Chat.ID, "Неверное кодовое слово, попробуйте еще раз.")
	}
}

func (b *Bot) handleStatistics(msg *tgbotapi.Message) {
	stats, _, err := b.svc.GetStatistics()
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Ошибка получения статистики.")
		log.Printf("Ошибка получения статистики: %v", err)
		return
	}

	subscribedUserCount, err := b.svc.GetSubscribedUserCount()
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Ошибка получения количества подписанных пользователей.")
		log.Printf("Ошибка получения количества подписанных пользователей: %v", err)
		return
	}

	response := fmt.Sprintf("Всего подписанных пользователей: %d\n", subscribedUserCount)
	for keyword, count := range stats {
		response += fmt.Sprintf("Кодовое слово: %s — использовано %d раз\n", keyword, count)
	}
	b.sendMessage(msg.Chat.ID, response)
}

func (b *Bot) handleAddFile(msg *tgbotapi.Message) {
	parts := strings.SplitN(msg.Text, " ", 2)
	if len(parts) != 2 {
		b.sendMessage(msg.Chat.ID, "Ошибка формата. Пожалуйста, отправьте кодовое слово и путь к файлу в формате: слово путь к файлу")
		return
	}

	keyword := parts[0]
	filePath := parts[1]

	exists, path := b.svc.KeywordExists(keyword)
	if exists {
		b.sendMessage(msg.Chat.ID, fmt.Sprintf("Слово уже существует: %s", path))
		return
	}

	err := b.svc.AddKeywordFile(keyword, filePath)
	if err != nil {
		b.sendMessage(msg.Chat.ID, fmt.Sprintf("Ошибка добавления файла: %v", err))
		return
	}

	b.sendMessage(msg.Chat.ID, fmt.Sprintf("Файл %s успешно добавлен для кодового слова %s.", filePath, keyword))
}

func (b *Bot) handleInfo(msg *tgbotapi.Message) {
	b.sendMessage(msg.Chat.ID, "Информация будет добавлена позже.")
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	b.api.Send(msg)
}

func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) sendFile(chatID int64, filePath string) {
	file := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filePath))
	b.api.Send(file)
}
