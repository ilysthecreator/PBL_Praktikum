package helper

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

var ErrInvalidCursor = errors.New("cursor tidak valid")

// EncodeCursor mengubah penanda menjadi satu string yang aman di URL.
func EncodeCursor(createdAt time.Time, id int) string {
	raw := strconv.FormatInt(createdAt.UTC().UnixNano(), 10) +
		"|" + strconv.Itoa(id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor membaca kembali penanda dari string.
func DecodeCursor(encoded string) (model.Cursor, error) {
	if strings.TrimSpace(encoded) == "" {
		return model.Cursor{}, ErrInvalidCursor
	}
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return model.Cursor{}, ErrInvalidCursor
	}
	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil || id < 1 {
		return model.Cursor{}, ErrInvalidCursor
	}
	return model.Cursor{CreatedAt: time.Unix(0, nanos).UTC(), ID: id}, nil
}

// ParseCursorQuery membaca query limit, cursor, search, grade, dan is_active
func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	limit := c.QueryInt("limit", 10)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	rawCursor := strings.TrimSpace(c.Query("cursor"))
	var after *model.Cursor
	if rawCursor != "" {
		cObj, err := DecodeCursor(rawCursor)
		if err != nil {
			return model.CursorQuery{}, BadRequest("cursor tidak valid")
		}
		after = &cObj
	}

	q := model.CursorQuery{
		Limit:  limit,
		Cursor: rawCursor,
		After:  after,
		Search: strings.TrimSpace(c.Query("search")),
		Grade:  strings.TrimSpace(c.Query("grade")),
	}

	if rawActive := c.Query("is_active"); rawActive != "" {
		val, err := strconv.ParseBool(rawActive)
		if err != nil {
			return model.CursorQuery{}, BadRequest("parameter is_active harus boolean")
		}
		q.IsActive = &val
	}

	return q, nil
}
