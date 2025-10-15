package dto

type SignUpRequest struct {
	Login       string `json:"login" binding:"required,min=3,max=128,alphanum"`
	Password    string `json:"password" binding:"required,min=6"`
	IsModerator bool   `json:"is_moderator"`
}

type SignInRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Login    *string `json:"login" binding:"omitempty,min=3,max=128,alphanum"`
	Password *string `json:"password" binding:"omitempty,min=6"`
}

type UserResponse struct {
	UserID      uint   `json:"user_id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}
