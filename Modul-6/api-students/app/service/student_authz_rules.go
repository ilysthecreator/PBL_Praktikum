package service

import (
	"api-students/app/model"
	"api-students/helper"
)

// CanAccessStudent memutuskan apakah seseorang berhak mengakses atau mengubah data mahasiswa tertentu.
//
// Aturan akses:
// 1. Pemilik data (ownerID sama dengan user ID pemanggil) selalu diizinkan.
// 2. Jika bukan pemilik, dicek apakah role pemanggil memiliki permission scope :any yang sesuai.
//
// Fungsi murni ini tidak mengimpor fiber maupun repository sehingga mudah diuji (unit testing).
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == ownerID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}
