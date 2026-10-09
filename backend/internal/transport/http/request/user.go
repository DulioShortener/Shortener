package request

type UserCreate struct {
	Username    string  `json:"username" validate:"required,username"`
	DisplayName *string `json:"display_name" validate:"omitempty,notblank,min=2,max=32"`
	Password    string  `json:"password" validate:"required,min=8,max=128,haslower,hasupper,hasdigit,hasspecial"`
}
