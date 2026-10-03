package model

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUserInfo struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int          `json:"expires_in"` // detik
	User        AuthUserInfo `json:"user"`
}

type StudentProfileInfo struct {
	NIM      string  `json:"nim"`
	Nama     string  `json:"nama"`
	Prodi    string  `json:"prodi"`
	Angkatan int     `json:"angkatan"`
	IPK      float64 `json:"ipk_terakhir,omitempty"`
}

type UserProfileResponse struct {
	ID      int                 `json:"id"`
	Email   string              `json:"email"`
	Role    string              `json:"role"`
	Student *StudentProfileInfo `json:"student,omitempty"`
}

type AuthClaims struct {
	UserID    int    `json:"user_id"`
	StudentID int    `json:"student_id,omitempty"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}
