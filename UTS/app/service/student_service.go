package service

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"UTS/app/model"
	"UTS/app/repository"
	"UTS/helper"
)

type StudentService struct {
	pool        *pgxpool.Pool
	studentRepo *repository.StudentRepository
	userRepo    *repository.UserRepository
}

func NewStudentService(
	pool *pgxpool.Pool,
	studentRepo *repository.StudentRepository,
	userRepo *repository.UserRepository,
) *StudentService {
	return &StudentService{
		pool:        pool,
		studentRepo: studentRepo,
		userRepo:    userRepo,
	}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}

	perPage := c.QueryInt("per_page", 10)
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 50 {
		perPage = 50
	}

	filter := model.StudentFilterQuery{
		Page:     page,
		PerPage:  perPage,
		Prodi:    c.Query("prodi"),
		Angkatan: c.QueryInt("angkatan", 0),
		Search:   c.Query("search"),
		Sort:     c.Query("sort"),
	}

	students, total, err := s.studentRepo.FindAll(c.UserContext(), filter)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	meta := &model.Meta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
	}

	return helper.SuccessList(c, "Data mahasiswa berhasil diambil", students, meta)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "format request tidak valid")
	}

	valErrors := make(map[string][]string)

	// Validasi NIM
	if strings.TrimSpace(req.NIM) == "" {
		valErrors["nim"] = append(valErrors["nim"], "NIM wajib diisi")
	} else if !helper.IsValidNIM(req.NIM) {
		valErrors["nim"] = append(valErrors["nim"], "NIM harus 12 digit angka")
	} else {
		exists, err := s.studentRepo.ExistsByNIM(c.UserContext(), req.NIM)
		if err != nil {
			return helper.Fail(c, fiber.StatusInternalServerError, "kesalahan server")
		}
		if exists {
			valErrors["nim"] = append(valErrors["nim"], "NIM sudah terdaftar")
		}
	}

	// Validasi Nama
	if strings.TrimSpace(req.Nama) == "" {
		valErrors["nama"] = append(valErrors["nama"], "nama wajib diisi")
	}

	// Validasi Email
	if strings.TrimSpace(req.Email) == "" {
		valErrors["email"] = append(valErrors["email"], "email wajib diisi")
	} else if !helper.IsValidEmail(req.Email) {
		valErrors["email"] = append(valErrors["email"], "Format email tidak valid")
	} else {
		exists, err := s.studentRepo.ExistsByEmail(c.UserContext(), req.Email)
		if err != nil {
			return helper.Fail(c, fiber.StatusInternalServerError, "kesalahan server")
		}
		if exists {
			valErrors["email"] = append(valErrors["email"], "email sudah terdaftar")
		}
	}

	// Validasi Prodi
	if strings.TrimSpace(req.Prodi) == "" {
		valErrors["prodi"] = append(valErrors["prodi"], "prodi wajib diisi")
	}

	// Validasi Angkatan
	if req.Angkatan == 0 {
		valErrors["angkatan"] = append(valErrors["angkatan"], "angkatan wajib diisi")
	} else if !helper.IsValidAngkatan(req.Angkatan) {
		valErrors["angkatan"] = append(valErrors["angkatan"], "angkatan harus 4 digit dan tidak melebihi tahun berjalan")
	}

	// Validasi IPK
	ipk := 0.00
	if req.IPKTerakhir != nil {
		if !helper.IsValidIPK(*req.IPKTerakhir) {
			valErrors["ipk_terakhir"] = append(valErrors["ipk_terakhir"], "ipk_terakhir harus bernilai antara 0.00 hingga 4.00")
		} else {
			ipk = *req.IPKTerakhir
		}
	}

	if len(valErrors) > 0 {
		return helper.FailValidation(c, valErrors)
	}

	// Password awal = NIM di-hash
	hashPassword, err := helper.HashPassword(req.NIM)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
	}

	// Buat dalam transaksi
	tx, err := s.pool.Begin(c.UserContext())
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memulai transaksi")
	}
	defer tx.Rollback(context.Background())

	// 1. Buat user
	userID, err := s.userRepo.Create(c.UserContext(), tx, req.Email, hashPassword, "mahasiswa")
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat user mahasiswa")
	}

	// 2. Buat student
	studentData := &model.Student{
		UserID:      userID,
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: ipk,
	}
	studentID, err := s.studentRepo.Create(c.UserContext(), tx, studentData)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat data mahasiswa")
	}

	if err := tx.Commit(c.UserContext()); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan transaksi")
	}

	respData := model.StudentItemResponse{
		ID:          studentID,
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: ipk,
	}

	return helper.Created(c, "Mahasiswa berhasil ditambahkan", respData)
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	claims, err := helper.GetAuthUser(c)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "unauthorized")
	}

	student, err := s.studentRepo.FindByID(c.UserContext(), id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan server")
	}
	if student == nil || student.DeletedAt != nil {
		return helper.Fail(c, fiber.StatusNotFound, "data mahasiswa tidak ditemukan")
	}

	// Pemeriksaan otorisasi: Mahasiswa hanya boleh melihat data sendiri
	if strings.EqualFold(claims.Role, "mahasiswa") && student.UserID != claims.UserID {
		return helper.Fail(c, fiber.StatusForbidden, "akses ditolak: tidak dapat mengakses data mahasiswa lain")
	}

	courses, err := s.studentRepo.GetStudentCourses(c.UserContext(), student.ID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mata kuliah mahasiswa")
	}

	totalSKS := 0
	for _, cr := range courses {
		totalSKS += cr.SKS
	}
	batasSKS := helper.HitungBatasSKS(student.IPKTerakhir)

	detail := model.StudentDetailResponse{
		ID:               student.ID,
		UserID:           student.UserID,
		NIM:              student.NIM,
		Nama:             student.Nama,
		Email:            student.Email,
		Prodi:            student.Prodi,
		Angkatan:         student.Angkatan,
		IPKTerakhir:      student.IPKTerakhir,
		DaftarMataKuliah: courses,
		TotalSKS:         totalSKS,
		BatasSKS:         batasSKS,
	}

	return helper.Success(c, fiber.StatusOK, "Detail mahasiswa berhasil diambil", detail)
}

func (s *StudentService) Update(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "format request tidak valid")
	}

	student, err := s.studentRepo.FindByID(c.UserContext(), id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan server")
	}
	if student == nil || student.DeletedAt != nil {
		return helper.Fail(c, fiber.StatusNotFound, "data mahasiswa tidak ditemukan")
	}

	valErrors := make(map[string][]string)

	if strings.TrimSpace(req.Nama) == "" {
		valErrors["nama"] = append(valErrors["nama"], "nama wajib diisi")
	}

	if strings.TrimSpace(req.Prodi) == "" {
		valErrors["prodi"] = append(valErrors["prodi"], "prodi wajib diisi")
	}

	if req.Angkatan == 0 {
		valErrors["angkatan"] = append(valErrors["angkatan"], "angkatan wajib diisi")
	} else if !helper.IsValidAngkatan(req.Angkatan) {
		valErrors["angkatan"] = append(valErrors["angkatan"], "angkatan harus 4 digit dan tidak melebihi tahun berjalan")
	}

	if req.IPKTerakhir != nil && !helper.IsValidIPK(*req.IPKTerakhir) {
		valErrors["ipk_terakhir"] = append(valErrors["ipk_terakhir"], "ipk_terakhir harus bernilai antara 0.00 hingga 4.00")
	}

	if len(valErrors) > 0 {
		return helper.FailValidation(c, valErrors)
	}

	if err := s.studentRepo.Update(c.UserContext(), id, req); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui data mahasiswa")
	}

	updatedStudent, err := s.studentRepo.FindByID(c.UserContext(), id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan server")
	}

	return helper.Success(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", updatedStudent)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	student, err := s.studentRepo.FindByID(c.UserContext(), id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan server")
	}
	if student == nil || student.DeletedAt != nil {
		return helper.Fail(c, fiber.StatusNotFound, "data mahasiswa tidak ditemukan")
	}

	if err := s.studentRepo.SoftDelete(c.UserContext(), id); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus mahasiswa")
	}

	return helper.NoContent(c)
}
