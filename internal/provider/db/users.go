package db

import (
	"database/sql"
	"geminiBackend/internal/domain"
	"time"
)

type UsersProvider struct {
	db *sql.DB
}

func NewUsersProvider(db *sql.DB) *UsersProvider {
	return &UsersProvider{db: db}
}

// Upsert пользователя по tg_id (создаёт или обновляет username, last_login)
func (p *UsersProvider) UpsertTelegramUser(tgID int, username string) error {
	now := time.Now().Format(time.RFC3339)
	_, err := p.db.Exec(`
				INSERT INTO users (tg_id, username, last_login)
        VALUES (?, ?, ?)
        ON CONFLICT(tg_id) DO UPDATE SET
          username=excluded.username,
					last_login=excluded.last_login,
          updated_at=CURRENT_TIMESTAMP
    `, tgID, username, now)
	return err
}

// GetUserByTelegramID возвращает пользователя по tg_id
func (p *UsersProvider) GetUserByTelegramID(tgID int) (*domain.UserDB, error) {
	row := p.db.QueryRow(`
		SELECT id, tg_id, username, gemini_api_key, is_admin, is_active, last_login, created_at, updated_at
		FROM users
		WHERE tg_id = ?
	`, tgID)
	var user domain.UserDB
	err := row.Scan(&user.ID, &user.TgID, &user.Username, &user.GeminiAPIKey, &user.IsAdmin, &user.IsActive, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SetGeminiAPIKey устанавливает или обновляет Gemini API ключ для пользователя по tg_id
func (p *UsersProvider) SetGeminiAPIKey(tgID int, apiKey string) error {
	_, err := p.db.Exec(`
		UPDATE users
		SET gemini_api_key = ?, updated_at = CURRENT_TIMESTAMP
		WHERE tg_id = ?
	`, apiKey, tgID)
	return err
}

// ClearGeminiAPIKey устанавливает gemini_api_key = NULL для пользователя
func (p *UsersProvider) ClearGeminiAPIKey(tgID int) error {
	_, err := p.db.Exec(`
		UPDATE users
		SET gemini_api_key = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE tg_id = ?
	`, tgID)
	return err
}

// SetAdmin устанавливает или обновляет статус администратора для пользователя по tg_id
func (p *UsersProvider) SetAdmin(tgID int, isAdmin bool) error {
	_, err := p.db.Exec(`
		UPDATE users
		SET is_admin = ?, updated_at = CURRENT_TIMESTAMP
		WHERE tg_id = ?
	`, isAdmin, tgID)
	return err
}

// SetActive устанавливает или обновляет статус активности для пользователя по tg_id
func (p *UsersProvider) SetActive(tgID int, isActive bool) error {
	_, err := p.db.Exec(`
		UPDATE users
		SET is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE tg_id = ?
	`, isActive, tgID)
	return err
}
