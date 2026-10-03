package service

import (
	"github.com/gofiber/fiber/v2"

	"UTS/app/model"
	"UTS/app/repository"
	"UTS/helper"
	"UTS/middleware"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	studentRepo *repository.StudentRepository
	jwtManager  *helper.JWTManager
}

func NewAuthService(
	userRepo *repository.UserRepository,
	studentRepo *repository.StudentRepository,
	jwtManager *helper.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		studentRepo: studentRepo,
		jwtManager:  jwtManager,
	}
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "format request tidak valid")
	}

	ip := c.IP()

	// Validasi input
	valErrors := make(map[string][]string)
	if req.Email == "" {
		valErrors["email"] = append(valErrors["email"], "email wajib diisi")
	} else if !helper.IsValidEmail(req.Email) {
		valErrors["email"] = append(valErrors["email"], "format email tidak valid")
	}

	if req.Password == "" {
		valErrors["password"] = append(valErrors["password"], "password wajib diisi")
	} else if len(req.Password) < 8 {
		valErrors["password"] = append(valErrors["password"], "password minimal 8 karakter")
	}

	if len(valErrors) > 0 {
		return helper.FailValidation(c, valErrors)
	}

	user, err := s.userRepo.FindByEmail(c.UserContext(), req.Email)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan server")
	}

	if user == nil || !helper.CheckPassword(user.Password, req.Password) {
		middleware.GlobalLoginTracker.RecordFailure(ip)
		return helper.Fail(c, fiber.StatusUnauthorized, "kredensial salah")
	}

	var studentID int
	if user.Role == "mahasiswa" {
		student, err := s.studentRepo.FindByUserID(c.UserContext(), user.ID)
		if err != nil {
			return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan server")
		}
		if student == nil || student.DeletedAt != nil {
			middleware.GlobalLoginTracker.RecordFailure(ip)
			return helper.Fail(c, fiber.StatusUnauthorized, "kredensial salah atau akun sudah dinonaktifkan")
		}
		studentID = student.ID
	}

	middleware.GlobalLoginTracker.RecordSuccess(ip)

	claims := model.AuthClaims{
		UserID:    user.ID,
		StudentID: studentID,
		Email:     user.Email,
		Role:      user.Role,
	}

	tokenStr, err := s.jwtManager.GenerateToken(claims)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token autentikasi")
	}

	resp := model.LoginResponse{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwtManager.GetAccessTTL().Seconds()),
		User: model.AuthUserInfo{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", resp)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	claims, err := helper.GetAuthUser(c)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "unauthorized")
	}

	user, err := s.userRepo.FindByID(c.UserContext(), claims.UserID)
	if err != nil || user == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "user tidak ditemukan")
	}

	resp := model.UserProfileResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	if user.Role == "mahasiswa" {
		student, err := s.studentRepo.FindByUserID(c.UserContext(), user.ID)
		if err != nil {
			return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan server")
		}
		if student != nil && student.DeletedAt == nil {
			resp.Student = &model.StudentProfileInfo{
				NIM:      student.NIM,
				Nama:     student.Nama,
				Prodi:    student.Prodi,
				Angkatan: student.Angkatan,
				IPK:      student.IPKTerakhir,
			}
		}
	}

	return helper.Success(c, fiber.StatusOK, "data profil berhasil diambil", resp)
}
