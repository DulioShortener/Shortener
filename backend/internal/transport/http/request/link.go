package request

type LinkCreate struct {
	URL string `json:"url" validate:"required,httpurl"`
}
