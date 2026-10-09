package handlers

import (
	"api-students/app/service"
)

type UserHandler = service.UserService
type StudentHandler = service.StudentService
type AuthHandler = service.AuthService
type NilaiHandler = service.NilaiService

var NewUserHandler = service.NewUserService
var NewStudentHandler = service.NewStudentService
var NewAuthHandler = service.NewAuthService
var NewNilaiHandler = service.NewNilaiService
