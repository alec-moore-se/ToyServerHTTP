package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if token == "" {
		return "", errors.New("no token")
	}

	key := strings.Split(token, " ")
	if len(key) != 2 {
		return "", errors.New("invalid token")
	}

	if key[0] != "ApiKey" {
		return "", errors.New("invalid auth header")
	}

	return key[1], nil

}
