package user

import (
	"database/sql"
	"encoding/json"
	"github.com/hitesh-goel/ekanek/internal/handlers/auth"
	"github.com/hitesh-goel/ekanek/internal/handlers/response"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

type jwtResponse struct {
	Jwt string `json:"jwt"`
}

// User Type
type CreateUser struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type GetUser struct {
	Id        string `json:"id" db:"uid"`
	FirstName string `json:"firstname" db:"first_name"`
	LastName  string `json:"lastname" db:"last_name"`
	Email     string `json:"email" db:"email"`
	Password  string `json:"password" db:"password"`
}

func getHash(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (user *CreateUser) isValidUser() bool {
	return user.Email != "" && user.Password != ""
}

func userSignup(w http.ResponseWriter, r *http.Request, db *sql.DB, key string) {
	var user CreateUser
	err := json.NewDecoder(r.Body).Decode(&user)

	if !user.isValidUser() || err != nil {
		response.RespondWithError(w, r, "pass valid user entry", http.StatusBadRequest)
		return
	}

	user.Password, err = getHash([]byte(user.Password))
	if err != nil {
		log.Println("Error while hashing the Password", err.Error())
		response.RespondWithError(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var query = `
		WITH uuid AS (
			SELECT * FROM uuid_generate_v1mc()
		)
		INSERT INTO users (
			uid,
            first_name,
            last_name,
            email,
			password
        ) VALUES (
			(SELECT * FROM uuid),
            $1,
            $2,
            $3,
            $4
        ) RETURNING uid`
	uid := ""
	err = db.QueryRow(query, user.FirstName, user.LastName, user.Email, user.Password).Scan(&uid)

	if err != nil {
		log.Println("Error while saving user to database: ", err.Error())
		response.RespondWithError(w, r, "something went wrong", http.StatusInternalServerError)
		return
	}

	jwtToken, err := auth.GenerateJWT(uid, key)
	if err != nil {
		log.Println("Error generating jwtToken", err.Error())
		response.RespondWithError(w, r, "something went wrong", http.StatusInternalServerError)
		return
	}
	resp := &jwtResponse{
		Jwt: jwtToken,
	}
	response.RespondWithSuccess(w, r, "success", resp, http.StatusOK)
}

func userLogin(w http.ResponseWriter, r *http.Request, db *sql.DB, key string) {
	var user CreateUser
	var dbUser GetUser
	err := json.NewDecoder(r.Body).Decode(&user)

	if !user.isValidUser() || err != nil {
		log.Println("Not a valid user record", err.Error())
		response.RespondWithError(w, r, "pass valid user entry", http.StatusBadRequest)
		return
	}

	var query = `select uid, password from users where email = $1`
	row := db.QueryRow(query, user.Email)
	_ = row.Scan(&dbUser.Id, &dbUser.Password)

	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)

	err = bcrypt.CompareHashAndPassword(dbPass, userPass)

	if err != nil {
		log.Println("Error while comparing passwords", err.Error())
		response.RespondWithError(w, r, "Wrong Password!", http.StatusForbidden)
		return
	}

	jwtToken, err := auth.GenerateJWT(dbUser.Id, key)
	if err != nil {
		log.Println("Error generating jwtToken", err.Error())
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := &jwtResponse{
		Jwt: jwtToken,
	}
	response.RespondWithSuccess(w, r, "success", resp, http.StatusOK)
}

func HandleSignup(key string, db *sql.DB) (string, func(http.ResponseWriter, *http.Request)) {
	return "/api/v1/user/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Println("Wrong request method")
			response.RespondWithError(w, r, "wrong http method", http.StatusMethodNotAllowed)
			return
		}
		userSignup(w, r, db, key)
	}
}

func HandleLogin(key string, db *sql.DB) (string, func(http.ResponseWriter, *http.Request)) {
	return "/api/v1/user/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			log.Println("Wrong request method")
			response.RespondWithError(w, r, "wrong http method", http.StatusMethodNotAllowed)
			return
		}
		userLogin(w, r, db, key)
	}
}
