package dto

type AuthResponse struct {
	Message      string       `json:"message"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	User         UserResponse `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AccessTokenResponse struct {
	Message     string `json:"message"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

type SignUpRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type SignInRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	UserID      uint   `json:"user_id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}

type UpdateUserRequest struct {
	Login    *string `json:"login,omitempty"`
	Password *string `json:"password,omitempty"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
