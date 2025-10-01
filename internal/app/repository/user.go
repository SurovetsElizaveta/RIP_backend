package repository

import (
	"errors"
	"rip/internal/app/ds"

	"gorm.io/gorm"
)

func (r *Repository) CreateUser(user *ds.User) error {
	if user.Login == "" {
		return errors.New("логин не может быть пустым")
	}
	if len(user.Password) < 6 {
		return errors.New("пароль должен содержать минимум 6 символов")
	}

	return r.db.Create(user).Error
}

func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByID(userID uint) (*ds.User, error) {
	var user ds.User
	err := r.db.First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUser(userID uint, updates map[string]interface{}) error {
	var user ds.User
	if err := r.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("пользователь не найден")
		}
		return err
	}

	return r.db.Model(&user).Updates(updates).Error
}

func (r *Repository) DeleteUser(userID uint) error {
	result := r.db.Delete(&ds.User{}, userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("пользователь не найден")
	}
	return nil
}
