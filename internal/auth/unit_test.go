package auth

import (
	"testing"
	"github.com/google/uuid"
	"time"
)

func TestCreationOfToken(t *testing.T) {

	userId, _ := uuid.NewRandom()
	secret := "Hello Ayush!"
	date_exp := 24 * time.Hour 

	token, err := MakeJWT(userId, secret, date_exp)
	if err != nil {t.Error("")}
	gotUser, err := ValidateJWT(token, secret)
	if err != nil {t.Error("")}
	
	if gotUser != userId {
		t.Error("User not equal")
	}
}