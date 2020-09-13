package auth

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

const (
	authHeaderKey = "Authorization"
)

var (
	authHeaderValueRegex = regexp.MustCompile("[[:space:]]([[:alnum:]]|[[:punct:]])+")
	errNoJWTFound        = errors.New("no JWT found for user")
)

func GetUserId(r *http.Request) (string, error) {
	jwt, err := extractJWT(r)
	if err != nil {
		return "", err
	}

	claims, err := VerifyJwt(jwt, "")
	if err != nil {
		return "", err
	}

	if claims.UserID == "" {
		return "", errors.New("no valid user id found")
	}

	return claims.UserID, nil
}

func extractJWT(req *http.Request) (string, error) {
	jwt := strings.TrimSpace(authHeaderValueRegex.FindString(req.Header.Get(authHeaderKey)))
	if jwt != "" {
		return jwt, nil
	}
	return "", errNoJWTFound
}
