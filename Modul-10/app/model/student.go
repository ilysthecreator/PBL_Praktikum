package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     string    `json:"grade"`
	IsActive  bool      `json:"is_active"`
	OwnerID   *int      `json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateStudentRequest struct {
	NIM   string `json:"nim" validate:"required,nim"`
	Name  string `json:"name" validate:"required,min=3,max=100"`
	Grade string `json:"grade" validate:"required,oneof=A B C D E"`
}
type ReplaceStudentRequest struct {
	NIM      string `json:"nim" validate:"required,nim"`
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Grade    string `json:"grade" validate:"required,oneof=A B C D E"`
	IsActive bool   `json:"is_active"`
}
type PatchStudentRequest struct {
	NIM      *string `json:"nim,omitempty" validate:"omitnil,nim"`
	Name     *string `json:"name,omitempty" validate:"omitnil,min=3,max=100"`
	Grade    *string `json:"grade,omitempty" validate:"omitnil,oneof=A B C D E"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type ListQuery struct {
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
	Sort     string `query:"sort"`
	Order    string `query:"order"`
	Search   string `query:"search"`
	IsActive *bool  `query:"is_active"`
	Grade    string `query:"grade"`
}

func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type WebResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type CursorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *CursorMeta `json:"meta,omitempty"`
}