package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingMethod := jwt.SigningMethodHS256
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("token is invalid")
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not get subject from token claims: %v", err)
	}

	userID, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Subject is not valid ID: %v", err)
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	var ErrNoBearerToken = errors.New("no bearer token found")
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoBearerToken
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && strings.ToLower(parts[0]) == "bearer") {
		return "", fmt.Errorf("invalid bearer token")
	}
	if len(parts[1]) == 0 {
		return "", fmt.Errorf("empty bearer token")
	}

	return parts[1], nil
}

func RefreshToken() (string, error) {
	// 32 bytes string
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}
	// hex encode random bytes to string
	refreshToken := hex.EncodeToString(randomBytes)
	return refreshToken, nil
}
