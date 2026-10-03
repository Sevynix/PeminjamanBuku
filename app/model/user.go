package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,strongpassword"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type AssignRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

type PatchUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitnil,min=3,max=30,username"`
	Email    *string `json:"email,omitempty" validate:"omitnil,email,max=255"`
}

type UserListQuery struct {
	ListQuery
	Role string
}