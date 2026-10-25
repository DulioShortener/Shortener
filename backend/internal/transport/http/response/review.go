package response

import (
	"strconv"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
	"github.com/DulioShortener/Shortener/backend/internal/service"
)

type ReviewComment struct {
	ID         string      `json:"id"`
	AuthorID   string      `json:"author_id"`
	ParentID   *string     `json:"parent_id,omitempty"`
	Body       *string     `json:"body"`
	Rating     *int        `json:"rating,omitempty"`
	CreatedAt  string      `json:"created_at"`
	HasReplies *bool       `json:"has_replies,omitempty"`
	Replies    *ReviewPage `json:"replies,omitempty"`
}

type ReviewPage struct {
	Items      []ReviewComment `json:"items"`
	HasMore    bool            `json:"has_more"`
	NextCursor *string         `json:"next_cursor"`
}

func ReviewCommentFromEntity(comment entity.ReviewComment, authenticated bool) ReviewComment {
	result := ReviewComment{
		ID: strconv.FormatInt(comment.ID, 10), CreatedAt: formatTime(comment.CreatedAt),
		AuthorID: strconv.FormatInt(comment.UserID, 10),
	}
	if comment.ParentID != nil {
		parentID := strconv.FormatInt(*comment.ParentID, 10)
		result.ParentID = &parentID
	}
	if comment.DeletedAt == nil {
		result.Body = &comment.Body
		result.Rating = comment.Rating
	}
	if authenticated {
		result.HasReplies = &comment.HasReplies
	}
	return result
}

func ReviewPageFromEntity(page entity.ReviewPage, authenticated bool) ReviewPage {
	result := ReviewPage{
		Items: make([]ReviewComment, 0, len(page.Items)), HasMore: page.HasMore,
	}
	for _, comment := range page.Items {
		result.Items = append(result.Items, ReviewCommentFromEntity(comment, authenticated))
	}
	if page.NextCursor != nil {
		value := strconv.FormatInt(page.NextCursor.CreatedAt.UnixMilli(), 10) + ":" +
			strconv.FormatInt(page.NextCursor.ID, 10)
		result.NextCursor = &value
	}
	return result
}

func ReviewListingFromService(listing service.ReviewListing, authenticated bool) ReviewPage {
	page := ReviewPageFromEntity(listing.Page, authenticated)
	if authenticated {
		for index := range page.Items {
			id := listing.Page.Items[index].ID
			preview := ReviewPageFromEntity(listing.Previews[id], true)
			page.Items[index].Replies = &preview
		}
	}
	return page
}

func ReviewDetailFromService(detail service.ReviewDetail, authenticated bool) ReviewComment {
	result := ReviewCommentFromEntity(detail.Comment, authenticated)
	if detail.Preview != nil {
		preview := ReviewPageFromEntity(*detail.Preview, true)
		result.Replies = &preview
	}
	return result
}
