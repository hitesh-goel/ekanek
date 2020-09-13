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

func getHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

func (user *CreateUser) isValidUser() bool {
	return user.Email != "" && user.Password != ""
}

func userSignup(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user CreateUser
	err := json.NewDecoder(r.Body).Decode(&user)

	if !user.isValidUser() || err != nil {
		response.RespondWithError(w, r, "pass valid user entry", http.StatusBadRequest)
		return
	}

	user.Password = getHash([]byte(user.Password))
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
        )`
	_, err = db.Exec(query, user.FirstName, user.LastName, user.Email, user.Password)
	if err != nil {
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	response.RespondWithStatus(w, r, "success", http.StatusOK)
}

func userLogin(w http.ResponseWriter, r *http.Request, privateKey string, db *sql.DB) {
	var user CreateUser
	var dbUser GetUser
	err := json.NewDecoder(r.Body).Decode(&user)

	if !user.isValidUser() || err != nil {
		response.RespondWithError(w, r, "pass valid user entry", http.StatusBadRequest)
		return
	}

	var query = `select uid, password from users where email = $1`
	row := db.QueryRow(query, user.Email)

	_ = row.Scan(&dbUser.Id, &dbUser.Password)

	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)

	passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)

	if passErr != nil {
		log.Println(passErr)
		response.RespondWithError(w, r, "Wrong Password!", http.StatusForbidden)
		return
	}

	jwtToken, err := auth.GenerateJWT(dbUser.Id, privateKey)
	if err != nil {
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := &jwtResponse{
		Jwt: jwtToken,
	}
	response.RespondWithStatus(w, r, resp, http.StatusOK)
}

func HandleSignup(db *sql.DB) (string, func(http.ResponseWriter, *http.Request)) {
	return "/user/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			response.RespondWithError(w, r, "wrong http method", http.StatusMethodNotAllowed)
			return
		}
		userSignup(w, r, db)
	}
}

func HandleLogin(privateKey string, db *sql.DB) (string, func(http.ResponseWriter, *http.Request)) {
	return "/user/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			response.RespondWithError(w, r, "wrong http method", http.StatusMethodNotAllowed)
			return
		}
		userLogin(w, r, privateKey, db)
	}
}
