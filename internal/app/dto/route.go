package dto

import (
	"errors"
)

type Route struct {
	RouteID     int    `json:"id"`
	Title       string `json:"title"`
	Distance    int    `json:"distance"`
	Description string `json:"description"`
	ImageUrl    string `json:"image_url"`
	Delay       int    `json:"delay"`
	Status      string `json:"status"`
}

type CreateRoute struct {
	Title       string `json:"title" binding:"required,min=1,max=128"`
	Distance    int    `json:"distance" binding:"required,min=1"`
	Description string `json:"description" binding:"max=1000"`
	Delay       int    `json:"delay" binding:"min=0"`
}

type UpdateRoute struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=128"`
	Distance    *int    `json:"distance" binding:"omitempty,min=1"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	Delay       *int    `json:"delay" binding:"omitempty,min=0"`
	Status      *string `json:"status" binding:"omitempty,oneof=активен неактивен"`
}

func (r *CreateRoute) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	if r.Distance <= 0 {
		return errors.New("distance must be positive")
	}
	return nil
}
