package auth

import "encoding/base64"

func ToAuthResponse(result AuthResult) AuthResponse {
	return AuthResponse{
		UserID:       result.UserID.String(),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		KDFSalt:      base64.StdEncoding.EncodeToString(result.KDFSalt),
	}
}
