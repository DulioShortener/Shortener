package request

type Login struct {
	Username string `json:"username" validate:"required,username"`
	Password string `json:"password" validate:"required,max=128"`
}
