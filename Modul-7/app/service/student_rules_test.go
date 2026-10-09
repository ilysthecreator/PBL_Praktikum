package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

func TestCountTotalPages(t *testing.T) {
	cases := []struct {
		total, limit, want int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
		{5, 0, 0},
	}

	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d",
				tc.total, tc.limit, tc.want, got)
		}
	}
}

func TestApplyStudentPatch(t *testing.T) {
	initial := model.Student{
		ID:       1,
		NIM:      "434241099",
		Name:     "Ahmad Dani",
		Grade:    "A",
		IsActive: true,
	}

	inactive := false
	newGrade := "B"
	result := ApplyStudentPatch(initial, model.PatchStudentRequest{
		IsActive: &inactive,
		Grade:    &newGrade,
	})

	if result.IsActive != false {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if result.Grade != "B" {
		t.Errorf("grade seharusnya 'B', dapat %s", result.Grade)
	}
	if result.NIM != "434241099" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestValidateCreateStudent_Declarative(t *testing.T) {
	// Kasus 1: Valid
	reqValid := model.CreateStudentRequest{
		NIM:   "434241099",
		Name:  "Muhammad Fadhil Ilyas",
		Grade: "A",
	}
	if errs := helper.ValidateStruct(reqValid); errs != nil {
		t.Errorf("seharusnya lolos validasi deklaratif, tetapi error: %v", errs)
	}

	// Kasus 2: Semua field kosong
	reqInvalid := model.CreateStudentRequest{}
	errs := helper.ValidateStruct(reqInvalid)
	if errs == nil {
		t.Fatal("seharusnya gagal validasi pada struct kosong")
	}
	if errs["nim"] != "wajib diisi" || errs["name"] != "wajib diisi" || errs["grade"] != "wajib diisi" {
		t.Errorf("pesan error validasi tidak sesuai: %v", errs)
	}

	// Kasus 3: Custom NIM validator (harus 9-18 digit angka)
	reqBadNIM := model.CreateStudentRequest{
		NIM:   "123abc78", // bukan angka & kurang dari 9 digit
		Name:  "Test User",
		Grade: "B",
	}
	errsNIM := helper.ValidateStruct(reqBadNIM)
	if errsNIM["nim"] != "format NIM tidak valid (harus 9-18 digit angka)" {
		t.Errorf("harus ada error format NIM: %v", errsNIM)
	}
}

func TestValidatePatchStudent_Omitnil(t *testing.T) {
	// D.2 butir 3: {"name":""} DITOLAK, sedangkan body tanpa name TETAP DITERIMA
	emptyName := ""
	reqEmptyName := model.PatchStudentRequest{
		Name: &emptyName,
	}
	errsEmpty := helper.ValidateStruct(reqEmptyName)
	if errsEmpty == nil || errsEmpty["name"] == "" {
		t.Fatalf("seharusnya ditolak ketika name bernilai string kosong: %v", errsEmpty)
	}

	// Body tanpa name (hanya grade) diterima
	grade := "A"
	reqOmittedName := model.PatchStudentRequest{
		Grade: &grade,
	}
	errsOmitted := helper.ValidateStruct(reqOmittedName)
	if errsOmitted != nil {
		t.Fatalf("seharusnya diterima ketika name di-omit: %v", errsOmitted)
	}
}

func TestIsEmptyStudentPatch(t *testing.T) {
	emptyReq := model.PatchStudentRequest{}
	if !IsEmptyStudentPatch(emptyReq) {
		t.Error("seharusnya true untuk Patch request kosong")
	}

	name := "Budi"
	nonEmptyReq := model.PatchStudentRequest{Name: &name}
	if IsEmptyStudentPatch(nonEmptyReq) {
		t.Error("seharusnya false jika ada field yang diisi")
	}
}
