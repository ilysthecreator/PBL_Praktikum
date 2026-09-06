package service

import (
	"testing"

	"api-students/app/model"
)

// Perhatikan: pengujian ini tidak menyalakan server, tidak menyentuh
// database, dan tidak membuat fiber.Ctx sama sekali.

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

func TestApplyPatch(t *testing.T) {
	initial := model.Student{
		ID:       1,
		NIM:      "220101001",
		Name:     "Ahmad Dani",
		Grade:    "A",
		IsActive: true,
	}

	// Kasus 1: Mengubah sebagian atribut secara valid
	inactive := false
	newGrade := "A+"
	result, errs := ApplyPatch(initial, model.PatchStudentRequest{
		IsActive: &inactive,
		Grade:    &newGrade,
	})

	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error validasi: %v", errs)
	}
	if result.IsActive != false {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if result.Grade != "A+" {
		t.Errorf("grade seharusnya 'A+', dapat %s", result.Grade)
	}
	if result.NIM != "220101001" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}

	// Kasus 2: Mengirim field string kosong (seharusnya menghasilkan error validasi)
	emptyName := "   "
	_, errsEmpty := ApplyPatch(initial, model.PatchStudentRequest{
		Name: &emptyName,
	})
	if errsEmpty["name"] != "tidak boleh kosong" {
		t.Errorf("diharapkan error 'tidak boleh kosong', dapat %v", errsEmpty["name"])
	}
}

func TestValidateCreate(t *testing.T) {
	// Kasus 1: Valid
	reqValid := model.CreateStudentRequest{
		NIM:   "220101002",
		Name:  "Siti Aminah",
		Grade: "A",
	}
	if errs := ValidateCreate(reqValid); len(errs) != 0 {
		t.Errorf("seharusnya lolos validasi, tetapi error: %v", errs)
	}

	// Kasus 2: Semua field kosong
	reqInvalid := model.CreateStudentRequest{}
	errs := ValidateCreate(reqInvalid)
	if len(errs) != 3 {
		t.Errorf("seharusnya terdapat 3 error field, dapat %d: %v", len(errs), errs)
	}
	if errs["nim"] != "wajib diisi" || errs["name"] != "wajib diisi" || errs["grade"] != "wajib diisi" {
		t.Errorf("pesan error validasi tidak sesuai: %v", errs)
	}
}

func TestValidateReplace(t *testing.T) {
	reqValid := model.ReplaceStudentRequest{
		NIM:      "220101001",
		Name:     "Ahmad Dani Pratama",
		Grade:    "A+",
		IsActive: true,
	}
	if errs := ValidateReplace(reqValid); len(errs) != 0 {
		t.Errorf("seharusnya lolos validasi PUT, tetapi error: %v", errs)
	}

	reqInvalid := model.ReplaceStudentRequest{
		NIM: "  ",
	}
	errs := ValidateReplace(reqInvalid)
	if errs["nim"] != "wajib diisi pada PUT" || errs["name"] != "wajib diisi pada PUT" {
		t.Errorf("pesan error validasi PUT tidak sesuai: %v", errs)
	}
}

func TestIsEmptyPatch(t *testing.T) {
	emptyReq := model.PatchStudentRequest{}
	if !IsEmptyPatch(emptyReq) {
		t.Error("seharusnya true untuk Patch request kosong")
	}

	name := "Budi"
	nonEmptyReq := model.PatchStudentRequest{Name: &name}
	if IsEmptyPatch(nonEmptyReq) {
		t.Error("seharusnya false jika ada field yang diisi")
	}
}
