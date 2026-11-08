package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

type SupervisorService struct {
	supervisors repository.SupervisorRepository
	students    repository.StudentRepository
	tx          *repository.TxManager
}

func NewSupervisorService(
	supervisors repository.SupervisorRepository,
	students repository.StudentRepository,
	tx *repository.TxManager,
) *SupervisorService {
	return &SupervisorService{
		supervisors: supervisors,
		students:    students,
		tx:          tx,
	}
}

// ---------- GET /supervisors ----------
func (s *SupervisorService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	list, err := s.supervisors.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar pembimbing berhasil diambil", list)
}

// ---------- GET /supervisors/:id ----------
func (s *SupervisorService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	sup, err := s.supervisors.FindByID(ctx, id)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "pembimbing ditemukan", sup)
}

// ---------- POST /supervisors/:id/assign ----------
// Menetapkan dosen pembimbing tugas akhir ke mahasiswa.
// Dijaga permission supervision:manage.
func (s *SupervisorService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.AssignSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var updatedSupervisor model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// 1. Kunci baris student untuk mencegah mahasiswa ditetapkan ke dua pembimbing secara serentak
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// 2. Ambil kuota pembimbing secara atomik
		var err error
		updatedSupervisor, err = repo.IncrementSupervised(ctx, supervisorID)
		if err != nil {
			return err
		}

		// 3. Catat supervisi. Bila sudah punya pembimbing, UNIQUE constraint di DB membatalkan langkah 2 lewat rollback
		return repo.InsertSupervision(ctx, supervisorID, req.StudentID)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Created(c, "berhasil menetapkan pembimbing", updatedSupervisor,
		"/api/v1/supervisors/"+strconv.Itoa(supervisorID))
}

// ---------- POST /supervisions/switch ----------
// Mengganti pembimbing tugas akhir secara atomik dan bebas deadlock.
func (s *SupervisorService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.SwitchSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if req.FromSupervisorID == req.ToSupervisorID {
		return helper.Validation(map[string]string{
			"to_supervisor_id": "tidak boleh sama dengan from_supervisor_id",
		})
	}

	var newSupervisor model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// Kunci mahasiswa
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// Kunci KEDUA pembimbing dalam urutan id menaik untuk mencegah deadlock
		if err := repo.LockSupervisors(ctx, req.FromSupervisorID, req.ToSupervisorID); err != nil {
			return err
		}

		// Hapus bimbingan lama
		if err := repo.DeleteSupervision(ctx, req.FromSupervisorID, req.StudentID); err != nil {
			return err
		}

		// Kurangi bimbingan dosen lama
		if _, err := repo.DecrementSupervised(ctx, req.FromSupervisorID); err != nil {
			return err
		}

		// Tambah bimbingan dosen baru (gagal bila penuh, me-rollback seluruh operasi)
		var err error
		newSupervisor, err = repo.IncrementSupervised(ctx, req.ToSupervisorID)
		if err != nil {
			return err
		}

		// Catat supervisi baru
		return repo.InsertSupervision(ctx, req.ToSupervisorID, req.StudentID)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "berhasil mengganti pembimbing", newSupervisor)
}

// ---------- PATCH /supervisors/:id ----------
// Ubah kuota pembimbing dengan optimistic locking.
func (s *SupervisorService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateSupervisorQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.supervisors.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "kuota pembimbing berhasil diubah", updated)
}

func translateSupervisorError(err error) error {
	switch {
	case errors.Is(err, repository.ErrSupervisorNotFound):
		return helper.NotFound("pembimbing tidak ditemukan")
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrSupervisorFull):
		return helper.NewError(fiber.StatusConflict, "SUPERVISOR_FULL", "kuota pembimbing sudah penuh")
	case errors.Is(err, repository.ErrAlreadySupervised):
		return helper.NewError(fiber.StatusConflict, "ALREADY_SUPERVISED", "mahasiswa sudah memiliki pembimbing")
	case errors.Is(err, repository.ErrNotSupervised):
		return helper.NewError(fiber.StatusConflict, "NOT_SUPERVISED", "mahasiswa tidak dibimbing oleh dosen tersebut")
	case errors.Is(err, repository.ErrVersionConflict):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT", "data pembimbing sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrQuotaBelowFilled):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_SUPERVISED", "kuota tidak boleh lebih kecil dari jumlah mahasiswa yang dibimbing saat ini")
	default:
		return helper.Internal(err)
	}
}
