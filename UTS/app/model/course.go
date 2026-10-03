package model

import "time"

type Course struct {
	ID        int       `json:"id"`
	KodeMK    string    `json:"kode_mk"`
	NamaMK    string    `json:"nama_mk"`
	SKS       int       `json:"sks"`
	Semester  int       `json:"semester"`
	Kuota     int       `json:"kuota"`
	Terisi    int       `json:"terisi"`
	SisaKuota int       `json:"sisa_kuota"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type CourseDetail struct {
	ID            int       `json:"id"`
	EnrollmentID  int       `json:"enrollment_id"`
	KodeMK        string    `json:"kode_mk"`
	NamaMK        string    `json:"nama_mk"`
	SKS           int       `json:"sks"`
	Semester      int       `json:"semester"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}

type CourseFilterQuery struct {
	Semester  int    `query:"semester"`
	Search    string `query:"search"`
	Available bool   `query:"available"`
}
