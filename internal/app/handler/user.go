package handler

import (
	"fmt"
	"net/http"

	"rip/internal/app/ds"
	"rip/internal/app/dto"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) SignUp(ctx *gin.Context) {
	var request dto.SignUpRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	existingUser, err := h.Repository.GetUserByLogin(request.Login)
	if err == nil && existingUser != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("пользователь с логином %s уже существует", request.Login))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Errorf("Ошибка хеширования пароля: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка создания пользователя"))
		return
	}

	user := &ds.User{
		Login:       request.Login,
		Password:    string(hashedPassword),
		IsModerator: request.IsModerator,
	}

	if err := h.Repository.CreateUser(user); err != nil {
		logrus.Errorf("Ошибка создания пользователя: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка создания пользователя"))
		return
	}

	response := dto.UserResponse{
		UserID:      user.UserID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Пользователь успешно зарегистрирован",
		"user":    response,
	})
}

func (h *Handler) SignIn(ctx *gin.Context) {
	var request dto.SignInRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByLogin(request.Login)
	if err != nil || user == nil {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("неверный логин или пароль"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("неверный логин или пароль"))
		return
	}

	response := dto.UserResponse{
		UserID:      user.UserID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Успешный вход в систему",
		"user":    response,
	})
}

func (h *Handler) SignOut(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Успешный выход из системы",
	})
}

func (h *Handler) GetCurrentUser(ctx *gin.Context) {
	currentUserID := uint(1)

	user, err := h.Repository.GetUserByID(currentUserID)
	if err != nil || user == nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("пользователь не найден"))
		return
	}

	response := dto.UserResponse{
		UserID:      user.UserID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) UpdateUser(ctx *gin.Context) {
	var request dto.UpdateUserRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	currentUserID := uint(1)

	updates := make(map[string]interface{})

	if request.Login != nil {
		existingUser, err := h.Repository.GetUserByLogin(*request.Login)
		if err == nil && existingUser != nil && existingUser.UserID != uint(currentUserID) {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("логин %s уже занят", *request.Login))
			return
		}
		updates["login"] = *request.Login
	}

	if request.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
		if err != nil {
			logrus.Errorf("Ошибка хеширования пароля: %v", err)
			h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка обновления пользователя"))
			return
		}
		updates["password"] = string(hashedPassword)
	}

	if len(updates) == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("нет полей для обновления"))
		return
	}

	if err := h.Repository.UpdateUser(currentUserID, updates); err != nil {
		logrus.Errorf("Ошибка обновления пользователя: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка обновления пользователя"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Данные пользователя успешно обновлены",
	})
}
