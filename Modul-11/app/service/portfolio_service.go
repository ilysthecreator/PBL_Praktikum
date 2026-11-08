package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentRepository,
) *PortfolioService {
	return &PortfolioService{
		portfolios: portfolios,
		students:   students,
	}
}

// GET /students/:id/portfolio
func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// Validasi mahasiswa harus terdaftar di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	portfolio, err := s.portfolios.GetByStudentID(ctx, studentID)
	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio mahasiswa berhasil diambil", portfolio)
}

// PUT /students/:id/portfolio
func (s *PortfolioService) Upsert(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// Validasi mahasiswa harus terdaftar di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	var req model.UpsertPortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Headline = strings.TrimSpace(req.Headline)
	req.Bio = strings.TrimSpace(req.Bio)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var certs []model.Certificate
	if req.Certificates != nil {
		certs = make([]model.Certificate, 0, len(req.Certificates))
		for _, cr := range req.Certificates {
			certs = append(certs, model.Certificate{
				Name:   strings.TrimSpace(cr.Name),
				Issuer: strings.TrimSpace(cr.Issuer),
				Date:   strings.TrimSpace(cr.Date),
			})
		}
	}

	updated, err := s.portfolios.Upsert(ctx, studentID, req.Headline, req.Bio, certs)
	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio berhasil disimpan", updated)
}

// POST /students/:id/portfolio/projects
func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// Validasi mahasiswa harus terdaftar di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	var req model.AddProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Link = strings.TrimSpace(req.Link)
	req.Technologies = NormalizeTags(req.Technologies)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.portfolios.AddProject(ctx, studentID, model.Project{
		Title:        req.Title,
		Technologies: req.Technologies,
		Link:         req.Link,
		Year:         req.Year,
	})
	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Success(c, fiber.StatusCreated, "proyek berhasil ditambahkan", updated)
}

// DELETE /students/:id/portfolio/projects/:projectId
func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	projectID := c.Params("projectId")
	if projectID == "" {
		return helper.BadRequest("id proyek tidak boleh kosong")
	}

	// Validasi mahasiswa harus terdaftar di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if err := s.portfolios.DeleteProject(ctx, studentID, projectID); err != nil {
		return translatePortfolioError(err)
	}

	return helper.NoContent(c)
}

// GET /portfolios/stats/technologies
func (s *PortfolioService) GetTopTechnologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	stats, err := s.portfolios.GetTopTechnologies(ctx)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "statistik teknologi berhasil diambil", stats)
}

func translatePortfolioError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("format id tidak valid")
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data tidak ditemukan")
	case errors.Is(err, repository.ErrProjectLimit):
		return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT",
			"jumlah proyek sudah mencapai batas maksimal 20")
	default:
		return helper.Internal(err)
	}
}
