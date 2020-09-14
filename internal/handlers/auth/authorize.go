package auth

import (
	"context"
	"errors"
	"github.com/hitesh-goel/ekanek/internal/handlers/response"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	ctxUIDkey     = "uid"
	authHeaderKey = "Authorization"
)

var (
	authHeaderValueRegex = regexp.MustCompile("[[:space:]]([[:alnum:]]|[[:punct:]])+")
)

// Auth ...
func Auth(p string, h func(http.ResponseWriter, *http.Request)) (string, func(http.ResponseWriter, *http.Request)) {
	return p, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userId, err := Verify(r)
		if err != nil {
			log.Println("Auth Error", err.Error())
			response.RespondWithError(w, r, err.Error(), http.StatusUnauthorized)
			return
		}
		r = r.WithContext(WithUID(ctx, userId))
		h(w, r)
	}
}

func Verify(r *http.Request) (string, error) {
	jwt, err := extractJWT(r)
	if err != nil {
		return "", err
	}

	claims, err := VerifyJwt(jwt, os.Getenv("PRIVATE_KEY"))
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
	return "", errors.New("no JWT found for user")
}

func GetUID(ctx context.Context) (string, error) {
	s, ok := ctx.Value(ctxUIDkey).(string)
	if !ok {
		return "", errors.New("UID not found in context")
	}
	return s, nil
}

// WithUID ...
func WithUID(ctx context.Context, UID string) context.Context {
	return context.WithValue(ctx, ctxUIDkey, UID)
}
