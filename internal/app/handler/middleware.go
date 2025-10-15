package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			h.errorHandler(ctx, 401, errors.New("authorization header required"))
			ctx.Abort()
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			h.errorHandler(ctx, 401, errors.New("invalid authorization header format"))
			ctx.Abort()
			return
		}

		tokenString := authHeader[7:]

		// Проверяем blacklist
		isBlacklisted, err := h.Redis.IsInBlacklist(ctx.Request.Context(), tokenString)
		if err != nil {
			h.errorHandler(ctx, 500, errors.New("internal server error"))
			ctx.Abort()
			return
		}
		if isBlacklisted {
			h.errorHandler(ctx, 401, errors.New("token is invalidated"))
			ctx.Abort()
			return
		}

		claims, err := h.JWTManager.ValidateToken(tokenString)
		if err != nil {
			h.errorHandler(ctx, 401, errors.New("invalid token"))
			ctx.Abort()
			return
		}

		if claims.IsRefresh {
			h.errorHandler(ctx, 401, errors.New("refresh token not allowed"))
			ctx.Abort()
			return
		}

		// Сохраняем данные пользователя в контекст
		ctx.Set("user_id", claims.UserID)
		ctx.Set("user_login", claims.Login)
		ctx.Set("is_moderator", claims.IsModerator)

		ctx.Next()
	}
}

// ModeratorMiddleware middleware для проверки прав модератора
func (h *Handler) ModeratorMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		isModerator, exists := ctx.Get("is_moderator")
		if !exists || !isModerator.(bool) {
			h.errorHandler(ctx, 403, errors.New("insufficient permissions"))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
