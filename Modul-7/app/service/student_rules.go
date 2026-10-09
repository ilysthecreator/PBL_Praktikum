package service

import (
	"strings"

	"api-students/app/model"
)

// ApplyStudentPatch menggabungkan field PATCH ke model Student yang ada.
// Pemeriksaan format/aturan field sudah dilakukan oleh validator secara deklaratif.
func ApplyStudentPatch(current model.Student, req model.PatchStudentRequest) model.Student {
	if req.NIM != nil {
		current.NIM = strings.TrimSpace(*req.NIM)
	}
	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}
	if req.Grade != nil {
		current.Grade = strings.TrimSpace(*req.Grade)
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current
}

// IsEmptyStudentPatch memeriksa apakah tidak ada satu pun atribut yang dikirim pada PATCH
func IsEmptyStudentPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// CountTotalPages menghitung pembagian halaman ke atas secara integer tanpa angka pecahan.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
