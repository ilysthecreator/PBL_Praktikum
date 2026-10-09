package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{
		repo:  repo,
		perms: perms,
	}
}

// ---------- GET /students (Cursor Pagination & Content Negotiation) ----------
func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// D.4: Content Negotiation (JSON dan CSV)
	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	// D.3: Keyset Pagination
	rows, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	if format == helper.FormatCSV {
		return helper.WriteStudentsCSV(c, rows)
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar mahasiswa berhasil diambil", rows, meta)
}

// ---------- GET /students/:id ----------
func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateStudentError(err, "mahasiswa")
	}

	ownerID := 0
	if student.OwnerID != nil {
		ownerID = *student.OwnerID
	}

	if !CanAccessStudent(current, ownerID, s.perms, "student:read:any") {
		return helper.Forbidden("tidak berhak mengakses data mahasiswa ini")
	}

	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diambil", student)
}

// ---------- POST /students ----------
func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)
	req.Grade = strings.TrimSpace(req.Grade)

	// D.2: Validasi deklaratif
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ownerID := current.UserID
	newStudent, err := s.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
		OwnerID:  &ownerID,
	})
	if err != nil {
		return translateStudentError(err, "mahasiswa")
	}

	return helper.Created(c, "mahasiswa berhasil dibuat", newStudent, "/api/v1/students/"+strconv.Itoa(newStudent.ID))
}

// ---------- PUT /students/:id ----------
func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateStudentError(err, "mahasiswa")
	}

	ownerID := 0
	if existing.OwnerID != nil {
		ownerID = *existing.OwnerID
	}

	if !CanAccessStudent(current, ownerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data mahasiswa ini")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	result, err := s.repo.Update(ctx, model.Student{
		ID:       id,
		NIM:      strings.TrimSpace(req.NIM),
		Name:     strings.TrimSpace(req.Name),
		Grade:    strings.TrimSpace(req.Grade),
		IsActive: req.IsActive,
		OwnerID:  existing.OwnerID,
	})
	if err != nil {
		return translateStudentError(err, "mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa berhasil diperbarui", result)
}

// ---------- PATCH /students/:id ----------
func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateStudentError(err, "mahasiswa")
	}

	ownerID := 0
	if existing.OwnerID != nil {
		ownerID = *existing.OwnerID
	}

	if !CanAccessStudent(current, ownerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data mahasiswa ini")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyStudentPatch(req) {
		return helper.BadRequest("setidaknya satu field harus dikirim")
	}

	// D.2: Validasi deklaratif dengan omitnil
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	merged := ApplyStudentPatch(existing, req)
	result, err := s.repo.Update(ctx, merged)
	if err != nil {
		return translateStudentError(err, "mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa berhasil diperbarui", result)
}

// ---------- DELETE /students/:id ----------
func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateStudentError(err, "mahasiswa")
	}

	return helper.NoContent(c)
}

func translateStudentError(err error, entity string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound(fmt.Sprintf("%s tidak ditemukan", entity))
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("NIM sudah terdaftar")
	default:
		return helper.Internal(err)
	}
}
