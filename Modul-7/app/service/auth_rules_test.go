package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

func TestRegisterValidation_Declarative(t *testing.T) {
	// Password kuat diterima (Perbaikan Bug 6)
	reqValid := model.RegisterRequest{
		Username: "fadhil123",
		Email:    "fadhil@unair.ac.id",
		Password: "PasswordAman2026",
	}
	if errs := helper.ValidateStruct(reqValid); errs != nil {
		t.Fatalf("seharusnya registrasi dengan password kuat diterima: %v", errs)
	}

	// Password terlalu pendek
	reqShort := model.RegisterRequest{
		Username: "fadhil123",
		Email:    "fadhil@unair.ac.id",
		Password: "abc",
	}
	errsShort := helper.ValidateStruct(reqShort)
	if errsShort["password"] != "minimal 8 karakter" {
		t.Errorf("diharapkan error 'minimal 8 karakter', dapat: %v", errsShort["password"])
	}

	// Password terlalu umum
	reqCommon := model.RegisterRequest{
		Username: "fadhil123",
		Email:    "fadhil@unair.ac.id",
		Password: "password123",
	}
	errsCommon := helper.ValidateStruct(reqCommon)
	if errsCommon["password"] != "password terlalu umum" {
		t.Errorf("diharapkan error 'password terlalu umum', dapat: %v", errsCommon["password"])
	}

	// Password tanpa angka
	reqNoDigit := model.RegisterRequest{
		Username: "fadhil123",
		Email:    "fadhil@unair.ac.id",
		Password: "tanpaangka",
	}
	errsNoDigit := helper.ValidateStruct(reqNoDigit)
	if errsNoDigit["password"] != "harus memuat huruf dan angka" {
		t.Errorf("diharapkan error 'harus memuat huruf dan angka', dapat: %v", errsNoDigit["password"])
	}

	// Username karakter terlarang
	reqBadUsername := model.RegisterRequest{
		Username: "a b!",
		Email:    "bukan-email",
		Password: "abc",
	}
	errsBad := helper.ValidateStruct(reqBadUsername)
	if errsBad["username"] != "hanya boleh huruf, angka, titik, dan garis bawah" {
		t.Errorf("diharapkan error username: %v", errsBad["username"])
	}
	if errsBad["email"] != "format email tidak valid" {
		t.Errorf("diharapkan error email: %v", errsBad["email"])
	}
}
