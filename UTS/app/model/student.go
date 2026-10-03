package model

import "time"

type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Email       string     `json:"email,omitempty"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type StudentItemResponse struct {
	ID          int     `json:"id"`
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim"`
	Nama        string   `json:"nama"`
	Email       string   `json:"email"`
	Prodi       string   `json:"prodi"`
	Angkatan    int      `json:"angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir"`
}

type UpdateStudentRequest struct {
	Nama        string   `json:"nama"`
	Prodi       string   `json:"prodi"`
	Angkatan    int      `json:"angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir"`
}

type StudentDetailResponse struct {
	ID               int            `json:"id"`
	UserID           int            `json:"user_id"`
	NIM              string         `json:"nim"`
	Nama             string         `json:"nama"`
	Email            string         `json:"email"`
	Prodi            string         `json:"prodi"`
	Angkatan         int            `json:"angkatan"`
	IPKTerakhir      float64        `json:"ipk_terakhir"`
	DaftarMataKuliah []CourseDetail `json:"daftar_mata_kuliah"`
	TotalSKS         int            `json:"total_sks"`
	BatasSKS         int            `json:"batas_sks"`
}

type StudentFilterQuery struct {
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
	Prodi    string `query:"prodi"`
	Angkatan int    `query:"angkatan"`
	Search   string `query:"search"`
	Sort     string `query:"sort"`
}
