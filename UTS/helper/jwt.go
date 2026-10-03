package helper

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"UTS/app/model"
)

var (
	ErrInvalidToken = errors.New("token tidak valid")
	ErrExpiredToken = errors.New("token kedaluwarsa")
)

type JWTClaims struct {
	UserID    int    `json:"user_id"`
	StudentID int    `json:"student_id,omitempty"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey     []byte
	issuer        string
	accessTTL     time.Duration
}

func NewJWTManager(secretKey string, issuer string, accessTTL time.Duration) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		accessTTL: accessTTL,
	}
}

func (m *JWTManager) GetAccessTTL() time.Duration {
	return m.accessTTL
}

func (m *JWTManager) GenerateToken(claims model.AuthClaims) (string, error) {
	now := time.Now()
	jwtClaims := JWTClaims{
		UserID:    claims.UserID,
		StudentID: claims.StudentID,
		Email:     claims.Email,
		Role:      claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", claims.UserID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString(m.secretKey)
}

func (m *JWTManager) Parse(tokenStr string) (*model.AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &model.AuthClaims{
		UserID:    claims.UserID,
		StudentID: claims.StudentID,
		Email:     claims.Email,
		Role:      claims.Role,
	}, nil
}
