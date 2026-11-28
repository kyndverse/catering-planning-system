package model

type LoginRequest struct {
	Email    string `validate:"required,email" form:"email"`
	Password string `validate:"required,min=8" form:"password"`
}

