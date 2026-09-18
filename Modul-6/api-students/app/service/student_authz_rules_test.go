package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

func TestCanAccessStudent(t *testing.T) {
	// Setup permissions
	// admin: student:read:any, student:update:any
	// staff: student:read:any
	// user: no permissions
	perms := helper.NewPermissionSet(map[string][]string{
		"admin": {"student:read:any", "student:update:any"},
		"staff": {"student:read:any"},
		"user":  {},
	})

	tests := []struct {
		name          string
		current       model.AuthUser
		ownerID       int
		permission    string
		expectedAllow bool
	}{
		{
			name:          "user biasa mengakses data miliknya sendiri (read)",
			current:       model.AuthUser{UserID: 10, Role: "user"},
			ownerID:       10,
			permission:    "student:read:any",
			expectedAllow: true,
		},
		{
			name:          "user biasa mengubah data miliknya sendiri (update)",
			current:       model.AuthUser{UserID: 10, Role: "user"},
			ownerID:       10,
			permission:    "student:update:any",
			expectedAllow: true,
		},
		{
			name:          "user biasa mengakses data orang lain (read)",
			current:       model.AuthUser{UserID: 10, Role: "user"},
			ownerID:       99,
			permission:    "student:read:any",
			expectedAllow: false,
		},
		{
			name:          "user biasa mengubah data orang lain (update)",
			current:       model.AuthUser{UserID: 10, Role: "user"},
			ownerID:       99,
			permission:    "student:update:any",
			expectedAllow: false,
		},
		{
			name:          "staff membaca data mahasiswa orang lain (memiliki student:read:any)",
			current:       model.AuthUser{UserID: 2, Role: "staff"},
			ownerID:       99,
			permission:    "student:read:any",
			expectedAllow: true,
		},
		{
			name:          "staff mengubah data mahasiswa orang lain (tidak punya student:update:any)",
			current:       model.AuthUser{UserID: 2, Role: "staff"},
			ownerID:       99,
			permission:    "student:update:any",
			expectedAllow: false,
		},
		{
			name:          "admin membaca data mahasiswa siapapun",
			current:       model.AuthUser{UserID: 1, Role: "admin"},
			ownerID:       99,
			permission:    "student:read:any",
			expectedAllow: true,
		},
		{
			name:          "admin mengubah data mahasiswa siapapun",
			current:       model.AuthUser{UserID: 1, Role: "admin"},
			ownerID:       99,
			permission:    "student:update:any",
			expectedAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanAccessStudent(tt.current, tt.ownerID, perms, tt.permission)
			if result != tt.expectedAllow {
				t.Errorf("CanAccessStudent() = %v, expected %v", result, tt.expectedAllow)
			}
		})
	}
}
