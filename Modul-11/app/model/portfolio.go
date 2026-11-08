package model

import "time"

// Portfolio merepresentasikan satu portofolio lengkap milik seorang mahasiswa.
// Model ini tetap bersih dari tipe driver MongoDB (tidak ada primitive.ObjectID)
// agar tidak mengikat layer service atau transport ke database tertentu.
type Portfolio struct {
	ID           string        `json:"id"`
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type Project struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	Link         string   `json:"link,omitempty"`
	Year         int      `json:"year"`
}

type Certificate struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Date   string `json:"date"`
}

type UpsertPortfolioRequest struct {
	Headline     string               `json:"headline" validate:"required,min=3,max=100"`
	Bio          string               `json:"bio" validate:"required,min=10,max=1000"`
	Certificates []CertificateRequest `json:"certificates,omitempty" validate:"omitempty,max=20,dive"`
}

type CertificateRequest struct {
	Name   string `json:"name" validate:"required,min=2,max=100"`
	Issuer string `json:"issuer" validate:"required,min=2,max=100"`
	Date   string `json:"date" validate:"required,min=4,max=20"`
}

type AddProjectRequest struct {
	Title        string   `json:"title" validate:"required,min=2,max=100"`
	Technologies []string `json:"technologies" validate:"required,min=1,max=10,dive,min=1,max=30"`
	Link         string   `json:"link,omitempty" validate:"omitempty,url"`
	Year         int      `json:"year" validate:"required,min=2000,max=2100"`
}

type TechnologyStat struct {
	Technology string `json:"technology"`
	Count      int    `json:"count"`
}
