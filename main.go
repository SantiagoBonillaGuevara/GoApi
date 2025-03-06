package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var secretKey = []byte("secreto_super_seguro")

// Usuario quemado en código
type User struct {
	Name     string
	Password string
	Role     string
}

var users = []User{
	{"userA", "userA", "A"},
	{"userB", "userB", "B"},
	{"userC", "userC", "C"},
}

// Estructura para recibir credenciales
type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Generar token JWT
func GenerateJWT(user User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.Name
	claims["roles"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()
	claims["iat"] = time.Now().Unix()
	claims["iss"] = "api-go"

	return token.SignedString(secretKey)
}

// Endpoint de login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	for _, user := range users {
		if user.Name == creds.Name && user.Password == creds.Password {
			token, err := GenerateJWT(user)
			if err != nil {
				http.Error(w, "Error generando token", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"token": token})
			return
		}
	}
	http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
}

// Middleware para autenticación y roles
func AuthMiddleware(role string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				http.Error(w, "Token requerido", http.StatusUnauthorized)
				return
			}

			tokenString = strings.TrimPrefix(tokenString, "Bearer ")

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return secretKey, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Token inválido "+tokenString+" esta mrd no lee bien parece", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Error en los claims", http.StatusUnauthorized)
				return
			}

			if claims["roles"] != role {
				http.Error(w, "Acceso denegado", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Endpoints protegidos
func AHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Solo el rol A puede acceder aquí"))
}

func BHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Solo el rol B puede acceder aquí"))
}

func CHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Solo el rol C puede acceder aquí"))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/users/login", LoginHandler).Methods("POST")
	r.Handle("/users/a", AuthMiddleware("A")(http.HandlerFunc(AHandler))).Methods("GET")
	r.Handle("/users/b", AuthMiddleware("B")(http.HandlerFunc(BHandler))).Methods("GET")
	r.Handle("/users/c", AuthMiddleware("C")(http.HandlerFunc(CHandler))).Methods("GET")

	fmt.Println("Servidor corriendo en :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}