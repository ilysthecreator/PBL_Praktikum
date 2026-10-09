package tests

import (
	"testing"

	"api-students/models"
	"api-students/validators"
)

func TestValidatorStrongPassword(t *testing.T) {
	// Password kuat harus lolos
	strongReq := models.RegisterRequest{
		Username: "fadhil123",
		Email:    "fadhil@unair.ac.id",
		Password: "PasswordKuat2026",
	}
	errs := validators.ValidateStruct(strongReq)
	if errs != nil {
		t.Fatalf("password kuat seharusnya lolos, tapi error: %v", errs)
	}

	// Password terlalu umum harus ditolak
	commonReq := models.RegisterRequest{
		Username: "fadhil123",
		Email:    "fadhil@unair.ac.id",
		Password: "password123",
	}
	errsCommon := validators.ValidateStruct(commonReq)
	if errsCommon == nil || errsCommon["password"] != "password terlalu umum" {
		t.Fatalf("password umum harus menghasilkan 'password terlalu umum', got: %v", errsCommon)
	}
}

func TestValidatorStudentOmitnil(t *testing.T) {
	// PATCH dengan nama kosong harus ditolak (D.2 butir 3)
	emptyName := ""
	patchEmpty := models.PatchStudentRequest{
		Name: &emptyName,
	}
	errs := validators.ValidateStruct(patchEmpty)
	if errs == nil || errs["name"] == "" {
		t.Fatalf("name string kosong harus ditolak, got: %v", errs)
	}

	// PATCH tanpa field name (diabaikan) harus lolos
	grade := "A"
	patchValid := models.PatchStudentRequest{
		Grade: &grade,
	}
	errsValid := validators.ValidateStruct(patchValid)
	if errsValid != nil {
		t.Fatalf("patch tanpa name harus lolos, got: %v", errsValid)
	}
}

func TestValidatorCustomNIM(t *testing.T) {
	validStudent := models.CreateStudentRequest{
		NIM:   "434241099",
		Name:  "Muhammad Fadhil Ilyas",
		Grade: "A",
	}
	if errs := validators.ValidateStruct(validStudent); errs != nil {
		t.Fatalf("NIM valid harus lolos, got: %v", errs)
	}

	invalidStudent := models.CreateStudentRequest{
		NIM:   "123abc",
		Name:  "Test Student",
		Grade: "A",
	}
	errs := validators.ValidateStruct(invalidStudent)
	if errs == nil || errs["nim"] != "format NIM tidak valid (harus 9-18 digit angka)" {
		t.Fatalf("NIM non-digit harus ditolak, got: %v", errs)
	}
}
