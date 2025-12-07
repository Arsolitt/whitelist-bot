package user

import (
	"encoding/json"
	"time"
)

func (u User) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID         ID         `json:"id"`
		TelegramID TelegramID `json:"telegram_id"`
		ChatID     ChatID     `json:"chat_id"`
		FirstName  FirstName  `json:"first_name"`
		LastName   LastName   `json:"last_name"`
		Username   Username   `json:"username"`
		CreatedAt  time.Time  `json:"created_at"`
		UpdatedAt  time.Time  `json:"updated_at"`
	}{
		ID:         u.id,
		TelegramID: u.telegramID,
		ChatID:     u.chatID,
		FirstName:  u.firstName,
		LastName:   u.lastName,
		Username:   u.username,
		CreatedAt:  u.createdAt,
		UpdatedAt:  u.updatedAt,
	})
}

func (u *User) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID         ID         `json:"id"`
		TelegramID TelegramID `json:"telegram_id"`
		ChatID     ChatID     `json:"chat_id"`
		FirstName  FirstName  `json:"first_name"`
		LastName   LastName   `json:"last_name"`
		Username   Username   `json:"username"`
		CreatedAt  time.Time  `json:"created_at"`
		UpdatedAt  time.Time  `json:"updated_at"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	user, err := NewBuilder().
		ID(aux.ID).
		TelegramID(aux.TelegramID).
		ChatID(aux.ChatID).
		FirstName(aux.FirstName).
		LastName(aux.LastName).
		Username(aux.Username).
		CreatedAt(aux.CreatedAt).
		UpdatedAt(aux.UpdatedAt).
		Build()
	if err != nil {
		return err
	}

	*u = user
	return nil
}
