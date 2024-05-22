package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func WithJWTAuth(handlerFunc http.HandlerFunc, store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// token'ı istekten al (Auth header)
		tokenString := GetTokenFromRequest(r)
		// belirteci doğrulayın
		token, err := validateJWT(tokenString)
		if err != nil {
			log.Println("kimlik doğrulama belirteci başarısız oldu")
			permissionDenied(w)
			return
		}

		if !token.Valid {
			log.Println("kimlik doğrulama belirteci başarısız oldu")
			permissionDenied(w)
			return
		}
		// token'dan kullanıcı kimliğini al

		claims := token.Claims.(jwt.MapClaims)
		userID := claims["userID"].(string)
		_, err = store.GetUserByID(userID)
		if err != nil {
			log.Println("Kullanıcı bulunamadı!")
			permissionDenied(w)
			return
		}
		// işleyici işlevini çağırın ve bitiş noktasına devam et
		handlerFunc(w, r)
	}
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusUnauthorized, ErrorResponse{
		Error: fmt.Errorf("İzin Rededildi").Error(),
	})
}

func GetTokenFromRequest(r *http.Request) string {
	tokenAuth := r.Header.Get("Authorization")
	tokenQuery := r.URL.Query().Get("token")

	if tokenAuth != "" {
		return tokenAuth
	}

	if tokenQuery != "" {
		return tokenQuery
	}
	return ""
}

func validateJWT(t string) (*jwt.Token, error) {
	secret := Envs.JWTSecret
	return jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Kabul edilmeyen imzalama yöntemi: %v", t.Header["alg"])
		}

		return []byte(secret), nil
	})
}

func HashPassword(pw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CreateJWT(secret []byte, userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": strconv.Itoa(int(userID)),
		"expiresAt": time.Now().Add(time.Hour * 24 * 120).Unix(),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}