package pagination

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var ErrInvalidCursor = errors.New("cursor tidak valid")

type Cursor struct {
	CreatedAt time.Time
	ID        int
}

type CursorQuery struct {
	Limit    int
	Cursor   string
	After    *Cursor
	Search   string
	Grade    string
	IsActive *bool
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// EncodeCursor mengubah penanda menjadi string URL-safe base64.
func EncodeCursor(createdAt time.Time, id int) string {
	raw := strconv.FormatInt(createdAt.UTC().UnixNano(), 10) +
		"|" + strconv.Itoa(id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor membaca kembali penanda dari string base64.
func DecodeCursor(encoded string) (Cursor, error) {
	if strings.TrimSpace(encoded) == "" {
		return Cursor{}, ErrInvalidCursor
	}
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return Cursor{}, ErrInvalidCursor
	}
	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil || id < 1 {
		return Cursor{}, ErrInvalidCursor
	}
	return Cursor{CreatedAt: time.Unix(0, nanos).UTC(), ID: id}, nil
}

// ParseCursorQuery membaca query string untuk pagination keyset.
func ParseCursorQuery(c *fiber.Ctx) (CursorQuery, error) {
	limit := c.QueryInt("limit", 10)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	rawCursor := strings.TrimSpace(c.Query("cursor"))
	var after *Cursor
	if rawCursor != "" {
		cObj, err := DecodeCursor(rawCursor)
		if err != nil {
			return CursorQuery{}, ErrInvalidCursor
		}
		after = &cObj
	}

	q := CursorQuery{
		Limit:  limit,
		Cursor: rawCursor,
		After:  after,
		Search: strings.TrimSpace(c.Query("search")),
		Grade:  strings.TrimSpace(c.Query("grade")),
	}

	if rawActive := c.Query("is_active"); rawActive != "" {
		val, err := strconv.ParseBool(rawActive)
		if err != nil {
			return CursorQuery{}, errors.New("parameter is_active harus bernilai boolean")
		}
		q.IsActive = &val
	}

	return q, nil
}
