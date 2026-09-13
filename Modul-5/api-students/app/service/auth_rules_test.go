package service

import (
	"testing"
)

func TestCheckPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected string
	}{
		{	name:	"password kurang dari 8 karakter",password: "pass1",expected: "minimal 8 karakter",},
		{
			name:     "password hanya huruf tanpa angka",password: "passwordtanpaangka",expected: "harus memuat huruf dan angka",
		},
		{
			name:     "password hanya angka tanpa huruf",password: "1234567890",expected: "harus memuat huruf dan angka",
		},
		{
			name:     "password ada di daftar umum / lemah",password: "password123",expected: "password terlalu umum",
		},
		{
			name:     "password kuat dan valid",password: "rahasia123",expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPasswordStrength(tt.password)
			if got != tt.expected {
				t.Errorf("CheckPasswordStrength(%q) = %q; want %q", tt.password, got, tt.expected)
			}
		})
	}
}
