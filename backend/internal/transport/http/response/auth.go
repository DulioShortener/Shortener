package response

import (
	"strconv"

	"github.com/DulioShortener/Shortener/backend/internal/service"
)

type Login struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	User      User   `json:"user"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
}

func LoginFromResult(result service.LoginResult) Login {
	return Login{
		ID: strconv.FormatInt(result.Session.ID, 10), Token: result.Token,
		TokenType: "Bearer", User: UserFromEntity(result.User),
		CreatedAt: formatTime(result.Session.CreatedAt),
		ExpiresAt: formatTime(result.Session.ExpiresAt),
	}
}
