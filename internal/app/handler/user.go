package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"rip/internal/app/ds"
	"rip/internal/app/dto"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// SignUp godoc
// @Summary User registration
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.SignUpRequest true "User registration data"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/signup [post]
func (h *Handler) SignUp(ctx *gin.Context) {
	var request dto.SignUpRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Валидация
	if request.Login == "" || request.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("логин и пароль обязательны"))
		return
	}

	if len(request.Password) < 6 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("пароль должен содержать минимум 6 символов"))
		return
	}

	// Проверяем существование пользователя
	existingUser, err := h.Repository.GetUserByLogin(request.Login)
	if err == nil && existingUser != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("пользователь с логином %s уже существует", request.Login))
		return
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Errorf("Ошибка хеширования пароля: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка создания пользователя"))
		return
	}

	// Создаем пользователя
	user := &ds.User{
		Login:    request.Login,
		Password: string(hashedPassword),
	}

	if err := h.Repository.CreateUser(user); err != nil {
		logrus.Errorf("Ошибка создания пользователя: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка создания пользователя"))
		return
	}

	// Генерируем токены
	tokenPair, err := h.JWTManager.GenerateTokenPair(user.UserID, user.Login, user.IsModerator)
	if err != nil {
		logrus.Errorf("Ошибка генерации токенов: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка генерации токенов"))
		return
	}

	// Сохраняем refresh token в Redis
	err = h.Redis.StoreRefreshToken(ctx.Request.Context(), user.UserID, tokenPair.RefreshToken, h.JWTManager.GetRefreshTokenTTL())
	if err != nil {
		logrus.Errorf("Ошибка сохранения refresh token: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка сохранения токена"))
		return
	}

	response := dto.AuthResponse{
		Message:      "Пользователь успешно зарегистрирован",
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User: dto.UserResponse{
			UserID:      user.UserID,
			Login:       user.Login,
			IsModerator: user.IsModerator,
		},
	}

	ctx.JSON(http.StatusCreated, response)
}

// SignIn godoc
// @Summary User authentication
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.SignInRequest true "User credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/signin [post]
func (h *Handler) SignIn(ctx *gin.Context) {
	var request dto.SignInRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Валидация
	if request.Login == "" || request.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("логин и пароль обязательны"))
		return
	}

	// Ищем пользователя
	user, err := h.Repository.GetUserByLogin(request.Login)
	if err != nil || user == nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("неверный логин или пароль user"))
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("неверный логин или пароль pass"))
		return
	}

	// Генерируем токены
	tokenPair, err := h.JWTManager.GenerateTokenPair(user.UserID, user.Login, user.IsModerator)
	if err != nil {
		logrus.Errorf("Ошибка генерации токенов: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка генерации токенов"))
		return
	}

	// Сохраняем refresh token в Redis
	err = h.Redis.StoreRefreshToken(ctx.Request.Context(), user.UserID, tokenPair.RefreshToken, h.JWTManager.GetRefreshTokenTTL())
	if err != nil {
		logrus.Errorf("Ошибка сохранения refresh token: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка сохранения токена"))
		return
	}

	response := dto.AuthResponse{
		Message:      "Успешный вход в систему",
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User: dto.UserResponse{
			UserID:      user.UserID,
			Login:       user.Login,
			IsModerator: user.IsModerator,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

// SignOut godoc
// @Summary User logout
// @Description Invalidate user tokens
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/signout [post]
func (h *Handler) SignOut(ctx *gin.Context) {
	// Получаем токен из заголовка
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("authorization header required"))
		return
	}

	tokenString := authHeader[7:]

	// Парсим токен чтобы получить expiration time
	claims, err := h.JWTManager.ParseToken(tokenString)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid token"))
		return
	}

	// Добавляем access token в blacklist
	expiresIn := time.Until(time.Unix(claims.ExpiresAt, 0))
	if expiresIn > 0 {
		err = h.Redis.AddToBlacklist(ctx.Request.Context(), tokenString, expiresIn)
		if err != nil {
			logrus.Errorf("Ошибка добавления токена в blacklist: %v", err)
			h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка выхода из системы"))
			return
		}
	}

	// Удаляем refresh token
	userID, exists := ctx.Get("user_id")
	if exists {
		err = h.Redis.DeleteRefreshToken(ctx.Request.Context(), userID.(uint))
		if err != nil {
			logrus.Errorf("Ошибка удаления refresh token: %v", err)
		}
	}

	ctx.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Успешный выход из системы",
	})
}

// RefreshToken godoc
// @Summary Refresh JWT tokens
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(ctx *gin.Context) {
	var request dto.RefreshTokenRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if request.RefreshToken == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("refresh token обязателен"))
		return
	}

	// Валидируем refresh token
	claims, err := h.JWTManager.ValidateToken(request.RefreshToken)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("invalid refresh token"))
		return
	}

	if !claims.IsRefresh {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("not a refresh token"))
		return
	}

	// Проверяем что refresh token совпадает с сохраненным в Redis
	storedToken, err := h.Redis.GetRefreshToken(ctx.Request.Context(), claims.UserID)
	if err != nil || storedToken != request.RefreshToken {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("refresh token not found or expired"))
		return
	}

	// Получаем данные пользователя
	user, err := h.Repository.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("user not found"))
		return
	}

	// Генерируем новые токены
	tokenPair, err := h.JWTManager.GenerateTokenPair(user.UserID, user.Login, user.IsModerator)
	if err != nil {
		logrus.Errorf("Ошибка генерации токенов: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка генерации токенов"))
		return
	}

	// Обновляем refresh token в Redis
	err = h.Redis.StoreRefreshToken(ctx.Request.Context(), user.UserID, tokenPair.RefreshToken, h.JWTManager.GetRefreshTokenTTL())
	if err != nil {
		logrus.Errorf("Ошибка сохранения refresh token: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка сохранения токена"))
		return
	}

	response := dto.AuthResponse{
		Message:      "Токены успешно обновлены",
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User: dto.UserResponse{
			UserID:      user.UserID,
			Login:       user.Login,
			IsModerator: user.IsModerator,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

// GetCurrentUser godoc
// @Summary Get current user profile
// @Description Get current authenticated user information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/me [get]
func (h *Handler) GetCurrentUser(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	user, err := h.Repository.GetUserByID(userID.(uint))
	if err != nil || user == nil {
		h.errorHandler(ctx, http.StatusNotFound, errors.New("пользователь не найден"))
		return
	}

	response := dto.UserResponse{
		UserID:      user.UserID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateUser godoc
// @Summary Update current user
// @Description Update current user information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateUserRequest true "User update data"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /users/me [put]
func (h *Handler) UpdateUser(ctx *gin.Context) {
	var request dto.UpdateUserRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	updates := make(map[string]interface{})

	if request.Login != nil {
		// Проверяем что логин не занят другим пользователем
		existingUser, err := h.Repository.GetUserByLogin(*request.Login)
		if err == nil && existingUser != nil && existingUser.UserID != userID.(uint) {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("логин %s уже занят", *request.Login))
			return
		}
		updates["login"] = *request.Login
	}

	if request.Password != nil {
		if len(*request.Password) < 6 {
			h.errorHandler(ctx, http.StatusBadRequest, errors.New("пароль должен содержать минимум 6 символов"))
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
		if err != nil {
			logrus.Errorf("Ошибка хеширования пароля: %v", err)
			h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка обновления пользователя"))
			return
		}
		updates["password"] = string(hashedPassword)
	}

	if len(updates) == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("нет полей для обновления"))
		return
	}

	if err := h.Repository.UpdateUser(userID.(uint), updates); err != nil {
		logrus.Errorf("Ошибка обновления пользователя: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка обновления пользователя"))
		return
	}

	ctx.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Данные пользователя успешно обновлены",
	})
}

// package handler

// import (
// 	"fmt"
// 	"net/http"

// 	"rip/internal/app/ds"
// 	"rip/internal/app/dto"

// 	"github.com/gin-gonic/gin"
// 	"github.com/sirupsen/logrus"
// 	"golang.org/x/crypto/bcrypt"
// )

// func (h *Handler) SignUp(ctx *gin.Context) {
// 	var request dto.SignUpRequest

// 	if err := ctx.ShouldBindJSON(&request); err != nil {
// 		h.errorHandler(ctx, http.StatusBadRequest, err)
// 		return
// 	}

// 	existingUser, err := h.Repository.GetUserByLogin(request.Login)
// 	if err == nil && existingUser != nil {
// 		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("пользователь с логином %s уже существует", request.Login))
// 		return
// 	}

// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		logrus.Errorf("Ошибка хеширования пароля: %v", err)
// 		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка создания пользователя"))
// 		return
// 	}

// 	user := &ds.User{
// 		Login:       request.Login,
// 		Password:    string(hashedPassword),
// 		IsModerator: request.IsModerator,
// 	}

// 	if err := h.Repository.CreateUser(user); err != nil {
// 		logrus.Errorf("Ошибка создания пользователя: %v", err)
// 		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка создания пользователя"))
// 		return
// 	}

// 	response := dto.UserResponse{
// 		UserID:      user.UserID,
// 		Login:       user.Login,
// 		IsModerator: user.IsModerator,
// 	}

// 	ctx.JSON(http.StatusCreated, gin.H{
// 		"message": "Пользователь успешно зарегистрирован",
// 		"user":    response,
// 	})
// }

// func (h *Handler) SignIn(ctx *gin.Context) {
// 	var request dto.SignInRequest

// 	if err := ctx.ShouldBindJSON(&request); err != nil {
// 		h.errorHandler(ctx, http.StatusBadRequest, err)
// 		return
// 	}

// 	user, err := h.Repository.GetUserByLogin(request.Login)
// 	if err != nil || user == nil {
// 		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("неверный логин или пароль"))
// 		return
// 	}

// 	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
// 		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("неверный логин или пароль"))
// 		return
// 	}

// 	response := dto.UserResponse{
// 		UserID:      user.UserID,
// 		Login:       user.Login,
// 		IsModerator: user.IsModerator,
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"message": "Успешный вход в систему",
// 		"user":    response,
// 	})
// }

// func (h *Handler) SignOut(ctx *gin.Context) {

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"message": "Успешный выход из системы",
// 	})
// }

// func (h *Handler) GetCurrentUser(ctx *gin.Context) {
// 	currentUserID := uint(1)

// 	user, err := h.Repository.GetUserByID(currentUserID)
// 	if err != nil || user == nil {
// 		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("пользователь не найден"))
// 		return
// 	}

// 	response := dto.UserResponse{
// 		UserID:      user.UserID,
// 		Login:       user.Login,
// 		IsModerator: user.IsModerator,
// 	}

// 	ctx.JSON(http.StatusOK, response)
// }

// func (h *Handler) UpdateUser(ctx *gin.Context) {
// 	var request dto.UpdateUserRequest

// 	if err := ctx.ShouldBindJSON(&request); err != nil {
// 		h.errorHandler(ctx, http.StatusBadRequest, err)
// 		return
// 	}

// 	currentUserID := uint(1)

// 	updates := make(map[string]interface{})

// 	if request.Login != nil {
// 		existingUser, err := h.Repository.GetUserByLogin(*request.Login)
// 		if err == nil && existingUser != nil && existingUser.UserID != uint(currentUserID) {
// 			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("логин %s уже занят", *request.Login))
// 			return
// 		}
// 		updates["login"] = *request.Login
// 	}

// 	if request.Password != nil {
// 		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
// 		if err != nil {
// 			logrus.Errorf("Ошибка хеширования пароля: %v", err)
// 			h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка обновления пользователя"))
// 			return
// 		}
// 		updates["password"] = string(hashedPassword)
// 	}

// 	if len(updates) == 0 {
// 		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("нет полей для обновления"))
// 		return
// 	}

// 	if err := h.Repository.UpdateUser(currentUserID, updates); err != nil {
// 		logrus.Errorf("Ошибка обновления пользователя: %v", err)
// 		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка обновления пользователя"))
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"message": "Данные пользователя успешно обновлены",
// 	})
// }
