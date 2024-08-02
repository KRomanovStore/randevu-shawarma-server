package users

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserView struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

// Constructor
func NewUserView(user User) UserView {
	return UserView{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
