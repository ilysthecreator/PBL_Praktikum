package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"UTS/app/model"
	"UTS/app/repository"
	"UTS/helper"
)

type EnrollmentService struct {
	pool           *pgxpool.Pool
	enrollmentRepo *repository.EnrollmentRepository
	courseRepo     *repository.CourseRepository
	studentRepo    *repository.StudentRepository
}

func NewEnrollmentService(
	pool *pgxpool.Pool,
	enrollmentRepo *repository.EnrollmentRepository,
	courseRepo *repository.CourseRepository,
	studentRepo *repository.StudentRepository,
) *EnrollmentService {
	return &EnrollmentService{
		pool:           pool,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
		studentRepo:    studentRepo,
	}
}

func (s *EnrollmentService) Create(c *fiber.Ctx) error {
	claims, err := helper.GetAuthUser(c)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "unauthorized")
	}

	if !strings.EqualFold(claims.Role, "mahasiswa") {
		return helper.Fail(c, fiber.StatusForbidden, "hanya mahasiswa yang dapat mengambil mata kuliah")
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "format request tidak valid")
	}

	valErrors := make(map[string][]string)
	if req.CourseID <= 0 {
		valErrors["course_id"] = append(valErrors["course_id"], "course_id wajib diisi")
	}
	if strings.TrimSpace(req.TahunAkademik) == "" {
		valErrors["tahun_akademik"] = append(valErrors["tahun_akademik"], "tahun_akademik wajib diisi")
	} else if !helper.IsValidTahunAkademik(req.TahunAkademik) {
		valErrors["tahun_akademik"] = append(valErrors["tahun_akademik"], "format tahun_akademik harus sesuai contoh '2026/2027-Ganjil' atau '2026/2027-Genap'")
	}

	if len(valErrors) > 0 {
		return helper.FailValidation(c, valErrors)
	}

	student, err := s.studentRepo.FindByUserID(c.UserContext(), claims.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "kesalahan server mengambil data mahasiswa")
	}
	if student == nil || student.DeletedAt != nil {
		return helper.Fail(c, fiber.StatusForbidden, "data mahasiswa tidak aktif atau tidak ditemukan")
	}

	// Mulai Transaksi dengan Row Locking
	tx, err := s.pool.Begin(c.UserContext())
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memulai transaksi")
	}
	defer tx.Rollback(context.Background())

	// 1. Ambil data course dengan lock FOR UPDATE
	course, err := s.courseRepo.FindByIDForUpdate(c.UserContext(), tx, req.CourseID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "kesalahan database")
	}
	if course == nil {
		return helper.Fail(c, fiber.StatusNotFound, "mata kuliah tidak ditemukan")
	}

	// 2. Cek duplikasi pengambilan MK pada tahun akademik yang sama
	isEnrolled, err := s.enrollmentRepo.IsEnrolledInTA(c.UserContext(), tx, student.ID, req.CourseID, req.TahunAkademik)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "kesalahan saat memeriksa riwayat mata kuliah")
	}
	if isEnrolled {
		return helper.Fail(c, fiber.StatusConflict, "mata kuliah sudah pernah diambil pada tahun akademik yang sama")
	}

	// 3. Cek kuota mata kuliah
	terisi, err := s.courseRepo.GetCurrentEnrolledCountInTx(c.UserContext(), tx, req.CourseID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "kesalahan saat memeriksa kuota mata kuliah")
	}
	if terisi >= course.Kuota {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "kuota mata kuliah sudah penuh")
	}

	// 4. Cek batas SKS sesuai IPK
	batasSKS := helper.HitungBatasSKS(student.IPKTerakhir)
	totalSKSTerdaftar, err := s.enrollmentRepo.GetTotalSKSInTA(c.UserContext(), tx, student.ID, req.TahunAkademik)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "kesalahan saat memeriksa total SKS terdaftar")
	}

	if totalSKSTerdaftar+course.SKS > batasSKS {
		sisaSKS := batasSKS - totalSKSTerdaftar
		if sisaSKS < 0 {
			sisaSKS = 0
		}
		errMsg := fmt.Sprintf("total SKS melebihi batas maksimum (%d SKS). Sisa SKS yang dapat diambil: %d SKS, SKS mata kuliah ini: %d SKS", batasSKS, sisaSKS, course.SKS)
		return helper.Fail(c, fiber.StatusUnprocessableEntity, errMsg)
	}

	// 5. Tambahkan enrollment
	enrollment, err := s.enrollmentRepo.Create(c.UserContext(), tx, student.ID, req.CourseID, req.TahunAkademik)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan pendaftaran mata kuliah")
	}

	if err := tx.Commit(c.UserContext()); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan transaksi pendaftaran")
	}

	resp := model.EnrollmentResponse{
		ID:            enrollment.ID,
		StudentID:     enrollment.StudentID,
		CourseID:      enrollment.CourseID,
		TahunAkademik: enrollment.TahunAkademik,
		CreatedAt:     enrollment.CreatedAt,
	}

	return helper.Created(c, "Mata kuliah berhasil diambil", resp)
}

func (s *EnrollmentService) Delete(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "ID enrollment tidak valid")
	}

	claims, err := helper.GetAuthUser(c)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "unauthorized")
	}

	if !strings.EqualFold(claims.Role, "mahasiswa") {
		return helper.Fail(c, fiber.StatusForbidden, "hanya mahasiswa yang dapat membatalkan mata kuliah")
	}

	student, err := s.studentRepo.FindByUserID(c.UserContext(), claims.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "kesalahan server")
	}
	if student == nil {
		return helper.Fail(c, fiber.StatusForbidden, "data mahasiswa tidak ditemukan")
	}

	enrollment, err := s.enrollmentRepo.FindByID(c.UserContext(), id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "kesalahan server")
	}
	if enrollment == nil {
		return helper.Fail(c, fiber.StatusNotFound, "data enrollment tidak ditemukan")
	}

	// Cek kepemilikan KRS
	if enrollment.StudentID != student.ID {
		return helper.Fail(c, fiber.StatusForbidden, "akses ditolak: enrollment milik mahasiswa lain")
	}

	if err := s.enrollmentRepo.Delete(c.UserContext(), id); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membatalkan mata kuliah")
	}

	return helper.NoContent(c)
}
