package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"FullStackApp01/common"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxPassLength = 72
)

var (
	log = logger.GetOrCreate("api")
)

// Storage defines the interface for persistence operations
type Storage interface {
	Close() error
	GetCounter() (uint64, error)
	IncrementCounter() (uint64, error)
	SaveUser(username, password, role string) error
	GetUser(username string) (*common.User, error)
	UpdatePassword(username, newPassword string) error
	ResetCounter() error
}

// Server holds dependencies for API handlers
type Server struct {
	store  Storage
	jwtKey []byte
}

// NewServer creates a new API server
func NewServer(store Storage, jwtKey []byte) *Server {
	return &Server{
		store:  store,
		jwtKey: jwtKey,
	}
}

// CounterResponse is the DTO for counter responses
type CounterResponse struct {
	Value uint64 `json:"value"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (s *Server) EnableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
}

func (s *Server) Authorized(w http.ResponseWriter, r *http.Request, roles []string, next func()) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &common.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	for _, role := range roles {
		if claims.Role == role {
			next()
			return
		}
	}

	http.Error(w, "Forbidden: Insufficient permissions", http.StatusForbidden)
}

// GetUserFromToken extracts the username from the Authorization header
func (s *Server) GetUserFromToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", http.ErrNoCookie
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &common.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtKey, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	return claims.Username, nil
}

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	s.EnableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds common.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(creds.Password) > maxPassLength {
		http.Error(w, fmt.Sprintf("Password too long (max %d characters)", maxPassLength), http.StatusBadRequest)
		return
	}

	// Default role is user
	err := s.store.SaveUser(creds.Username, creds.Password, "user")
	if err != nil {
		if strings.Contains(err.Error(), "user already exists") {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}
	log.Debug("User created successfully", "user", creds.Username)
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	s.EnableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds common.Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := s.store.GetUser(creds.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Hash, []byte(creds.Password))
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &common.Claims{
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "role": user.Role})
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	log.Debug("User logged in successfully", "user", creds.Username)
}

func (s *Server) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	s.EnableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, err := s.GetUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) > maxPassLength || len(req.OldPassword) > maxPassLength {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Password too long (max %d characters)", maxPassLength)})
		return
	}

	// Verify old password
	user, err := s.store.GetUser(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	err = bcrypt.CompareHashAndPassword(user.Hash, []byte(req.OldPassword))
	if err != nil {
		http.Error(w, "Invalid old password", http.StatusUnauthorized)
		return
	}

	err = s.store.UpdatePassword(username, req.NewPassword)
	if err != nil {
		http.Error(w, "Could not update password", http.StatusInternalServerError)
		return
	}

	log.Debug("User changed password", "user", username)

	w.WriteHeader(http.StatusOK)
}

func (s *Server) HandleCounter(w http.ResponseWriter, r *http.Request) {
	s.EnableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodGet {
		val, err := s.store.GetCounter()
		if err != nil {
			http.Error(w, "Failed to get counter", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(CounterResponse{Value: val})
		if err != nil {
			http.Error(w, "Failed to encode counter", http.StatusInternalServerError)
			return
		}

		log.Debug("counter query", "value", val)

		return
	}

	if r.Method == http.MethodPost {
		// Require at least "user" role (or admin)
		s.Authorized(w, r, []string{"user", "admin"}, func() {
			val, err := s.store.IncrementCounter()
			if err != nil {
				http.Error(w, "Failed to increment counter", http.StatusInternalServerError)
				return
			}

			err = json.NewEncoder(w).Encode(CounterResponse{Value: val})
			if err != nil {
				http.Error(w, "Failed to encode counter", http.StatusInternalServerError)
				return
			}

			log.Debug("counter incremented", "new value", val)
		})
		return
	}

	if r.Method == http.MethodDelete {
		// Require "admin" role
		s.Authorized(w, r, []string{"admin"}, func() {
			err := s.store.ResetCounter()
			if err != nil {
				http.Error(w, "Failed to reset counter", http.StatusInternalServerError)
				return
			}

			err = json.NewEncoder(w).Encode(CounterResponse{Value: 0})
			if err != nil {
				http.Error(w, "Failed to encode counter", http.StatusInternalServerError)
				return
			}

			log.Debug("counter reset", "new value", 0)
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
