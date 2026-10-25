package request

import "encoding/json"

type ReviewCreate struct {
	Body   string `json:"body" validate:"required,notblank,max=2000"`
	Rating *int   `json:"rating" validate:"required,min=1,max=5"`
}

type ReplyCreate struct {
	Body   string          `json:"body" validate:"required,notblank,max=2000"`
	Rating json.RawMessage `json:"rating"`
}
