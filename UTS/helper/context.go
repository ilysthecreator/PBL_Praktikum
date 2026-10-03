package helper

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"UTS/app/model"
)

const LocalsAuthUser = "authUser"

var ErrNoAuthContext = errors.New("tidak ada konteks otentikasi")

func GetAuthUser(c *fiber.Ctx) (*model.AuthClaims, error) {
	val := c.Locals(LocalsAuthUser)
	if val == nil {
		return nil, ErrNoAuthContext
	}
	claims, ok := val.(*model.AuthClaims)
	if !ok {
		return nil, ErrNoAuthContext
	}
	return claims, nil
}
