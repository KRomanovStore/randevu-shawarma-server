package users

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

var db *sql.DB

// SetDatabase sets the database connection
func SetDatabase(database *sql.DB) {
	db = database
}

// RegisterRoutes registers all user routes
func RegisterRoutes(router *httprouter.Router) {
	router.POST("/users/login", Login)

	router.GET("/users", Authenticate(GetUsers))
	router.GET("/users/:id", Authenticate(GetUser))
	router.POST("/users", Authenticate(CreateUser))
	router.PUT("/users/:id", Authenticate(UpdateUser))
	router.DELETE("/users/:id", Authenticate(DeleteUser))

}

func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u := UserView{}
	var hashedPassword string
	err = db.QueryRow("SELECT id, name, email, password FROM public.\"Users\" WHERE email = $1", credentials.Email).Scan(&u.ID, &u.Name, &u.Email, &hashedPassword)
	if err == sql.ErrNoRows || !checkPasswordHash(credentials.Password, hashedPassword) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(credentials.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		//Domain:   "localhost",
	})

	json.NewEncoder(w).Encode(u)
}

func GetUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	email := r.Context().Value("email").(string)

	row := db.QueryRow("SELECT id, name, email FROM public.\"Users\" WHERE email = $1", email)
	var u UserView
	err := row.Scan(&u.ID, &u.Name, &u.Email)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(u)
}

func GetUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	row := db.QueryRow("SELECT id, name, email FROM public.\"Users\" WHERE id = $1", id)

	var u UserView
	err := row.Scan(&u.ID, &u.Name, &u.Email)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(u)
}

func CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var u User

	// Check Content-Type header
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type header is not application/json", http.StatusUnsupportedMediaType)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isValidEmail(u.Email) {
		http.Error(w, "Invalid Email", http.StatusBadRequest)
		return
	}
	if !isValidPassword(u.Password) {
		http.Error(w, "Invalid Password", http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashPassword(u.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u.Password = hashedPassword
	u.CreatedAt = time.Now()

	err = db.QueryRow("INSERT INTO public.\"Users\" (name, email, password, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		u.Name, u.Email, u.Password, u.CreatedAt).Scan(&u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(u)
}

func UpdateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashPassword(u.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u.Password = hashedPassword

	_, err = db.Exec("UPDATE public.\"Users\" SET name = $1, email = $2, password = $3 WHERE id = $4", u.Name, u.Email, u.Password, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(u)
}

func DeleteUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	_, err := db.Exec("DELETE FROM public.\"Users\" WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
