package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     string    `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateStudentRequest struct {
	NIM   string `json:"nim"`
	Name  string `json:"name"`
	Grade string `json:"grade"`
}

type ReplaceStudentRequest struct {
	NIM      string `json:"nim"`
	Name     string `json:"name"`
	Grade    string `json:"grade"`
	IsActive bool   `json:"is_active"`
}

type PatchStudentRequest struct {
	NIM      *string `json:"nim"`
	Name     *string `json:"name"`
	Grade    *string `json:"grade"`
	IsActive *bool   `json:"is_active"`
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
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}