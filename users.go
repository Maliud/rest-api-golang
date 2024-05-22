package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)


var errEmailRequired = errors.New("email zorunludur")
var errFirstNameRequired = errors.New("İsim zorunludur")
var errLastNameRequired = errors.New("Soyisim zorunludur")
var errPasswordRequired = errors.New("Şifre zorunludur")




type UserService struct {
	store Store
}

func NewUserService(s Store) *UserService {
	return &UserService{store: s}
}

func (s *UserService) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/users/register", s.handleUserRegister).Methods("POST")
}

func (s *UserService) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w,"Body okunurken hata oluştu!", http.StatusBadRequest)
		return
	}	

	defer r.Body.Close()

	var payload *User
	err = json.Unmarshal(body, &payload)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Geçersiz payload isteği!"})
		return
	}
	if err := validateUserPayload(payload); err != nil {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	hashedPW, err := HashPassword(payload.Password)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Kullanıcı oluşturulurken hata oluştu!"})
		return
	}
	payload.Password = hashedPW

	u, err := s.store.CreateUser(payload)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Kullanıcı oluşturulurken hata oluştu!"})
		return
	}

	token, err := createAndSetAuthCookie(u.ID, w)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Session oluşturulurken hata oluştu!"})
		return
	}

	WriteJSON(w, http.StatusCreated, token) 
}

func validateUserPayload(user *User) error {
	if user.Email == "" {
		return errEmailRequired
	}
	if user.FirstName == "" {
		return errFirstNameRequired
	}
	if user.LastName == "" {
		return errLastNameRequired
	}
	if user.Password == "" {
		return errPasswordRequired
	}
	return nil
}

func createAndSetAuthCookie(id int64, w http.ResponseWriter) (string, error) {
	secret := []byte(Envs.JWTSecret)
	token, err := CreateJWT(secret, id)
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name: "Authorization",
		Value: token,
	})

	return token, nil
}