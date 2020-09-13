package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"strings"
	"time"
)

type JwtClaims struct {
	jwt.StandardClaims
	UserClaims
}

type UserClaims struct {
	UserID string `json:"uid,omitempty"`
}

// GenerateJWT function
func GenerateJWT(id string, privateKey string) (string, error) {
	clock := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JwtClaims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  clock.Unix(),
			ExpiresAt: clock.Add(time.Hour * time.Duration(21600)).Unix(), // expiration after 3 Months
		},
		UserClaims: UserClaims{
			UserID: id,
		},
	})
	tokenString, err := token.SignedString([]byte(privateKey))
	if err != nil {
		log.Println("Error in JWT token generation", err)
		return "", err
	}
	return tokenString, nil
}

func VerifyJwt(jwtToken string, privateKey string) (JwtClaims, error) {
	var jwtHeader = struct {
		Type string `json:"typ"`
		Alg  string `json:"alg"`
	}{"", ""}

	var claims JwtClaims

	parts := strings.Split(jwtToken, ".")
	numParts := len(parts)
	if numParts != 3 {
		return claims, errors.New(fmt.Sprintf("JWT token should have 3 base64-encoded parts, but got %d", numParts))
	}

	data, err := base64.RawStdEncoding.DecodeString(parts[0]) // parts[0] is base64 of jwt header
	if err != nil {
		return claims, errors.New(fmt.Sprint("Base64 decoding of JWT header failed: ", err.Error()))
	}

	if err := json.NewDecoder(bytes.NewBuffer(data)).Decode(&jwtHeader); err != nil {
		return claims, errors.New(fmt.Sprint("JSON decoding of JWT header failed: ", err.Error()))
	}

	if jwtHeader.Type != "JWT" {
		return claims, errors.New("JWT type is incorrect")
	}

	if jwtHeader.Alg == "HS256" {
		return claims, errors.New(fmt.Sprint("JWT uses unsupported signature algorithm: ", jwtHeader.Alg))
	}

	data, err = base64.RawStdEncoding.DecodeString(parts[1]) // parts[1] is base64 of jwt claims
	if err != nil {
		return claims, errors.New(fmt.Sprint("Base64 decoding of JWT claims failed: ", err.Error()))
	}

	if err := json.NewDecoder(bytes.NewBuffer(data)).Decode(&claims); err != nil {
		return claims, errors.New(fmt.Sprint("JSON decoding of JWT claims failed: ", err.Error()))
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return claims, errors.New(fmt.Sprint("JWT Expired at: ", time.Unix(claims.ExpiresAt, 0)))
	}

	_, err = jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(privateKey), nil
	})

	if err != nil {
		return claims, errors.New("BAD JWT Signature")
	}

	// TODO: Verify User has an active session.
	// 1. Store the User specific logout timestamp in redis
	// 2. Check timestamp of issued jwt should be greater than the last logout time
	// 3. If not that that jwt is not a valid jwt

	return claims, nil
}
