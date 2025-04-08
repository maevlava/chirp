package auth_test

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maevlava/chirpy/internal/auth"
	"testing"
	"time"
)

func TestMakeJWT(t *testing.T) {
	userId := uuid.New()
	secret := "rahasisaYangSangatKuat"
	duration := 1 * time.Hour

	tokenString, err := auth.MakeJWT(userId, secret, duration)
	if err != nil {
		t.Error(err)
		return
	}
	if tokenString == "" {
		t.Errorf("MakeJWT() returned empty token string")
	}
}

func TestValidateJWT(t *testing.T) {
	secret := "rahasisaYangSangatKuatLainya"
	userId := uuid.New()
	validDuration := 1 * time.Hour

	t.Run("Valid token", func(t *testing.T) {
		tokenString, _ := auth.MakeJWT(userId, secret, validDuration)
		parsedUserId, err := auth.ValidateJWT(tokenString, secret)

		if err != nil {
			t.Errorf("ValidateJWT() with valid token error = %v, wantErr nil", err)
		}
		if parsedUserId != userId {
			t.Errorf("ValidateJWT() with valid token got UserID = %v, want %v", parsedUserId, userId)
		}
	})
	t.Run("InvalidSignature", func(t *testing.T) {
		tokenstring, _ := auth.MakeJWT(userId, secret, validDuration)
		wrongSecret := "RahasiaYangSalah"
		_, err := auth.ValidateJWT(tokenstring, wrongSecret)
		if err == nil {
			t.Errorf("ValidateJWT() with wrong secret error = nil, wantErr")
		}
		if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			t.Errorf("ValidateJWT() with wrong secret error type = %T, want specific invalid signature error (%v)", err, jwt.ErrTokenSignatureInvalid)
		}
	})
}
