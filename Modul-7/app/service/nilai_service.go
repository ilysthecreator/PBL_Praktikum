package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type NilaiService struct {
	nilaiRepo   repository.NilaiRepository
	studentRepo repository.StudentRepository
}

func NewNilaiService(
	nilaiRepo repository.NilaiRepository,
	studentRepo repository.StudentRepository,
) *NilaiService {
	return &NilaiService{
		nilaiRepo:   nilaiRepo,
		studentRepo: studentRepo,
	}
}

// GetByNIM mengambil daftar nilai dari mahasiswa berdasarkan NIM yang dikirim.
func (s *NilaiService) GetByNIM(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	nim := strings.TrimSpace(c.Params("nim"))
	if nim == "" {
		nim = strings.TrimSpace(c.Query("nim"))
	}

	if nim == "" {
		return helper.BadRequest("parameter NIM wajib diisi")
	}

	// 1. Cari data mahasiswa terlebih dahulu berdasarkan NIM
	student, err := s.studentRepo.FindByNIM(ctx, nim)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa dengan NIM tersebut tidak ditemukan")
		}
		return helper.Internal(err)
	}

	// 2. Ambil daftar nilai milik mahasiswa tersebut
	daftarNilai, err := s.nilaiRepo.FindByStudentID(ctx, student.ID)
	if err != nil {
		return helper.Internal(err)
	}

	// 3. Kembalikan response terstruktur
	result := model.StudentNilaiResponse{
		Student:     student,
		DaftarNilai: daftarNilai,
	}

	return helper.Success(c, fiber.StatusOK, "daftar nilai mahasiswa berhasil diambil", result)
}

// Create menambahkan data nilai baru untuk mahasiswa.
func (s *NilaiService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateNilaiRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.NamaMataKuliah = strings.TrimSpace(req.NamaMataKuliah)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// Pastikan student ada di database
	if _, err := s.studentRepo.FindByID(ctx, req.IDStudent); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	created, err := s.nilaiRepo.Create(ctx, model.Nilai{
		NamaMataKuliah: req.NamaMataKuliah,
		Nilai:          req.Nilai,
		IDStudent:      req.IDStudent,
	})
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Created(c, "nilai berhasil ditambahkan", created, "")
}
