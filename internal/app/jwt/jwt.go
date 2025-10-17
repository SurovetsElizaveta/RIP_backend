package jwt

import (
	"errors"
	"time"

	"rip/internal/app/config"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserID      uint   `json:"user_id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
	IsRefresh   bool   `json:"is_refresh"`
	jwt.StandardClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Manager struct {
	secret          string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewManager(cfg config.JWTConfig) *Manager {
	return &Manager{
		secret:          cfg.Secret,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

func (m *Manager) GetAccessTokenTTL() time.Duration {
	return m.accessTokenTTL
}

func (m *Manager) GetRefreshTokenTTL() time.Duration {
	return m.refreshTokenTTL
}

// func (m *Manager) GenerateTokenPair(userID uint, login string, isModerator bool) (*TokenPair, error) {
// 	accessTokenClaims := Claims{
// 		UserID:      userID,
// 		Login:       login,
// 		IsModerator: isModerator,
// 		IsRefresh:   false,
// 		StandardClaims: jwt.StandardClaims{
// 			ExpiresAt: time.Now().Add(m.accessTokenTTL).Unix(),
// 			IssuedAt:  time.Now().Unix(),
// 			Issuer:    "rip-service",
// 		},
// 	}

// 	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
// 	accessTokenString, err := accessToken.SignedString([]byte(m.secret))
// 	if err != nil {
// 		return nil, err
// 	}

// 	refreshTokenClaims := Claims{
// 		UserID:      userID,
// 		Login:       login,
// 		IsModerator: isModerator,
// 		IsRefresh:   true,
// 		StandardClaims: jwt.StandardClaims{
// 			ExpiresAt: time.Now().Add(m.refreshTokenTTL).Unix(),
// 			IssuedAt:  time.Now().Unix(),
// 			Issuer:    "rip-service",
// 		},
// 	}

// 	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
// 	refreshTokenString, err := refreshToken.SignedString([]byte(m.secret))
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &TokenPair{
// 		AccessToken:  accessTokenString,
// 		RefreshToken: refreshTokenString,
// 		TokenType:    "Bearer",
// 		ExpiresIn:    int64(m.accessTokenTTL.Seconds()),
// 	}, nil
// }

func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	return m.ValidateToken(tokenString)
}

func (m *Manager) GenerateAccessToken(userID uint, login string, isModerator bool) (string, error) {
	claims := Claims{
		UserID:      userID,
		Login:       login,
		IsModerator: isModerator,
		IsRefresh:   false,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(m.accessTokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "rip-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

func (m *Manager) GenerateTokenPair(userID uint, login string, isModerator bool) (*TokenPair, error) {
	accessToken, err := m.GenerateAccessToken(userID, login, isModerator)
	if err != nil {
		return nil, err
	}

	refreshToken, err := m.generateRefreshToken(userID, login, isModerator)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(m.accessTokenTTL.Seconds()),
	}, nil
}

func (m *Manager) generateRefreshToken(userID uint, login string, isModerator bool) (string, error) {
	claims := Claims{
		UserID:      userID,
		Login:       login,
		IsModerator: isModerator,
		IsRefresh:   true,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(m.refreshTokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "rip-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}
