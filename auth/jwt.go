package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	jwt "github.com/dgrijalva/jwt-go"
)

const secret = " "
const defaultExpiretime = int64(600000)

type User struct {
}

type Claims struct {
	Id   string
	User string
	jwt.StandardClaims
}

func (c Claims) GetId() string {
	return c.Id
}

func (c Claims) GetUser() string {
	return c.User
}

func createJwtToken(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.ParseClaims{
		"Id":        User.GetId(),
		"Name":      User.GetName(),
		"ExpiresAt": time.Now().Unix() + defaultExpiretime,
	})
	tokenString, err := jwt.SignedString([]byte(secret))
	if err != nil {
		fmt.Println("Unable to return signed token string")
	}
	return tokenString, nil
}

func ValidateToken(tokenString string) (User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return Claims, err
	}
}
