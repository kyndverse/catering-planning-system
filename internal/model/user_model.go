package model

import "time"

type UserResponse struct {
	ID        int64
	Name      string
	Email     string
	Role      string
	CreatedAt time.Time
}

type CreateUserRequest struct {
	Name     string `form:"name"     validate:"required"`
	Email    string `form:"email"    validate:"required,email"`
	Password string `form:"password" validate:"required,min=8"`
	Role     string `form:"role"     validate:"required"`
}

type UpdateUserRequest struct {
	ID   int64
	Name string `form:"name"  validate:"required"`
	Role string `form:"role"  validate:"required"`
}

type ChangePasswordRequest struct {
	ID       int64
	Password string `form:"password" validate:"required,min=8"`
}

type DeleteUserRequest struct {
	ID int64
}
