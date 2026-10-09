package response

import (
	"strconv"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
)

type Link struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Code      string `json:"code"`
	TargetURL string `json:"target_url"`
	ShortURL  string `json:"short_url"`
	CreatedAt string `json:"created_at"`
}

type LinkList struct {
	Links []Link `json:"links"`
}

func LinkFromEntity(link entity.Link, shortURLBase string) Link {
	return Link{
		ID: strconv.FormatInt(link.ID, 10), UserID: strconv.FormatInt(link.UserID, 10),
		Code: link.Code, TargetURL: link.TargetURL,
		ShortURL:  shortURLBase + "/r/" + link.Code,
		CreatedAt: formatTime(link.CreatedAt),
	}
}

func LinksFromEntities(links []entity.Link, shortURLBase string) LinkList {
	responses := make([]Link, 0, len(links))
	for _, link := range links {
		responses = append(responses, LinkFromEntity(link, shortURLBase))
	}
	return LinkList{Links: responses}
}
