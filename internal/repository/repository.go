package repository

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	createTables(db)

	return &Repository{db: db}, nil
}

func createTables(db *sql.DB) {
	query := `
    CREATE TABLE IF NOT EXISTS files (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        keyword TEXT NOT NULL,
        file_path TEXT NOT NULL,
        count INTEGER DEFAULT 0
    );
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        contact TEXT NOT NULL,
        subscribed INTEGER NOT NULL,
        admin INTEGER NOT NULL
    );
    `
	db.Exec(query)
}

func (r *Repository) SaveContact(userID int, contact string) error {
	if r.UserExists(userID) {
		return nil // Пользователь уже существует, ничего не делаем
	}

	query := `INSERT INTO users (user_id, contact, subscribed, admin) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, userID, contact, 0, 0)
	return err
}

func (r *Repository) UserExists(userID int) bool {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = ?)`
	err := r.db.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func (r *Repository) CheckSubscription(userID int) (bool, error) {
	var subscribed int
	query := `SELECT subscribed FROM users WHERE user_id = ?`
	err := r.db.QueryRow(query, userID).Scan(&subscribed)
	if err != nil {
		return false, err
	}
	return subscribed == 1, nil
}

func (r *Repository) UpdateSubscription(userID int, subscribed bool) error {
	query := `UPDATE users SET subscribed = ? WHERE user_id = ?`
	_, err := r.db.Exec(query, subscribed, userID)
	return err
}

func (r *Repository) GetFilePathByKeyword(keyword string) (string, error) {
	var filePath string
	query := `SELECT file_path FROM files WHERE keyword = ?`
	err := r.db.QueryRow(query, keyword).Scan(&filePath)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

func (r *Repository) AddKeywordFile(keyword string, filePath string) error {
	query := `INSERT INTO files (keyword, file_path) VALUES (?, ?)`
	_, err := r.db.Exec(query, keyword, filePath)
	return err
}

func (r *Repository) KeywordExists(keyword string) (bool, string) {
	var filePath string
	query := `SELECT file_path FROM files WHERE keyword = ?`
	err := r.db.QueryRow(query, keyword).Scan(&filePath)
	if err != nil {
		return false, ""
	}
	return true, filePath
}

func (r *Repository) IsAdmin(userID int) (bool, error) {
	var admin int
	query := `SELECT admin FROM users WHERE user_id = ?`
	err := r.db.QueryRow(query, userID).Scan(&admin)
	if err != nil {
		return false, err
	}
	return admin == 1, nil
}

func (r *Repository) SetAdmin(userID int, admin bool) error {
	query := `UPDATE users SET admin = ? WHERE user_id = ?`
	_, err := r.db.Exec(query, admin, userID)
	return err
}

func (r *Repository) IncrementKeywordCount(keyword string) error {
	query := `UPDATE files SET count = count + 1 WHERE keyword = ?`
	_, err := r.db.Exec(query, keyword)
	return err
}

func (r *Repository) GetStatistics() (map[string]int, int, error) {
	stats := make(map[string]int)
	query := `SELECT keyword, count FROM files`
	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var totalUsers int
	for rows.Next() {
		var keyword string
		var count int
		err := rows.Scan(&keyword, &count)
		if err != nil {
			log.Printf("Ошибка сканирования строки: %v", err)
			return nil, 0, err
		}
		stats[keyword] = count
		totalUsers += count
	}

	if err = rows.Err(); err != nil {
		log.Printf("Ошибка итерации строк: %v", err)
		return nil, 0, err
	}

	return stats, totalUsers, nil
}

func (r *Repository) GetSubscribedUserCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE subscribed = 1`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return 0, err
	}
	return count, nil
}
