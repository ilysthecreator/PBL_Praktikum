package service

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"UTS/app/model"
	"UTS/app/repository"
	"UTS/helper"
)

type CourseService struct {
	courseRepo *repository.CourseRepository
}

func NewCourseService(courseRepo *repository.CourseRepository) *CourseService {
	return &CourseService{courseRepo: courseRepo}
}

func (s *CourseService) List(c *fiber.Ctx) error {
	semester := c.QueryInt("semester", 0)
	search := c.Query("search")
	availableStr := strings.ToLower(c.Query("available"))
	available := availableStr == "true" || availableStr == "1"

	filter := model.CourseFilterQuery{
		Semester:  semester,
		Search:    search,
		Available: available,
	}

	courses, err := s.courseRepo.FindAll(c.UserContext(), filter)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar mata kuliah")
	}

	return helper.Success(c, fiber.StatusOK, "Daftar mata kuliah berhasil diambil", courses)
}
