package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	var password string
	var err error
	password, err = HashPassword("password")
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}
	fmt.Println(password)
	match, err := CheckPasswordHash("password", password)
	if err != nil {
		t.Errorf("Error checking password: %v", err)
	}
	if !match {
		t.Errorf("Passwords do not match")
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, "secretKey", time.Hour)
	if err != nil {
		t.Errorf("Error making JWT: %v", err)
	}
	fmt.Println(token)
	userID2, err := ValidateJWT(token, "secretKey")
	if err != nil {
		t.Errorf("Error validating JWT: %v", err)
	}
	if userID != userID2 {
		t.Errorf("User IDs do not match")
	}
	fmt.Println(userID2)
}
