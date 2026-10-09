package response

import (
	"strconv"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
)

type User struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func UserFromEntity(user entity.User) User {
	return User{
		ID: strconv.FormatInt(user.ID, 10), Username: user.Username,
		DisplayName: user.DisplayName, CreatedAt: formatTime(user.CreatedAt),
		UpdatedAt: formatTime(user.UpdatedAt),
	}
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
