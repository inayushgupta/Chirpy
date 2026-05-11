package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	//"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicsugar"
)


func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password, hash string) (bool, error) {
	what, _ ,err := argon2id.CheckHash(password, hash)
	return what, err
}



func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	
	var claims *jwt.RegisteredClaims = &jwt.RegisteredClaims{
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Issuer:    "chirpy-access",
		Subject: userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(tokenSecret))
	return ss, err
} 

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	x, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})

	// 1. Check the error immediately!
	if err != nil {
		return uuid.Nil, err
	}

	// 2. Extract the subject from the claims
	userIdString, err := x.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	// 3. Use Parse instead of MustParse to avoid crashing on bad strings
	id, err := uuid.Parse(userIdString)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}


func GetBearerToken(headers http.Header) (string, error) {
	headerValue := headers.Get("Authorization")

	// header not supplied error
	if headerValue == "" {
		return "", errors.New("Header not supplied")
	}

	prefix := "Bearer"

	if !strings.HasPrefix(headerValue, prefix) {
		return "", errors.New("Prefix absent")
	}

	return strings.TrimSpace(strings.TrimPrefix(headerValue, prefix)), nil

}


func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key)
}


func GetAPIKey(headers http.Header) (string, error) {
	value := headers.Get("Authorization")

	if value == "" {
		return "", errors.New("No value in header")
	} // no key provided

	prefix := "ApiKey"
	if !strings.HasPrefix(value, prefix) {
		return "", errors.New("Prefix not ApiKey")
	}

	return strings.TrimSpace(strings.TrimPrefix(value, prefix)), nil

}
