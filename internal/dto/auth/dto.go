package auth

import "github.com/google/uuid"

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	KDFSalt      string `json:"kdf_salt"`
}

type AuthResult struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
	KDFSalt      []byte
}
