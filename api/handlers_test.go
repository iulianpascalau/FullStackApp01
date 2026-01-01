package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"FullStackApp01/common"
	"FullStackApp01/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testVersion = "v1.2.3"

var testKey = []byte("test_secret")

func setupServer(t *testing.T) *Server {
	t.Helper()
	store := mock.NewMockStorage()

	// Ensure admin exists
	err := store.SaveUser("admin", "admin123", "admin")
	require.NoError(t, err)

	return NewServer(store, testVersion, testKey)
}

func TestHandleRegister(t *testing.T) {
	s := setupServer(t)

	t.Run("should register new user", func(t *testing.T) {
		creds := common.Credentials{Username: "newuser", Password: "password"}
		body, _ := json.Marshal(creds)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		s.HandleRegister(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		// Verify user exists in store
		user, err := s.store.GetUser("newuser")
		assert.NoError(t, err)
		assert.Equal(t, "user", user.Role)
	})
}

func TestHandleLogin(t *testing.T) {
	s := setupServer(t)

	t.Run("should login with correct credentials", func(t *testing.T) {
		creds := common.Credentials{Username: "admin", Password: "admin123"}
		body, _ := json.Marshal(creds)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		s.HandleLogin(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]string
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp["token"])
		assert.Equal(t, "admin", resp["role"])
	})

	t.Run("should fail with invalid password", func(t *testing.T) {
		creds := common.Credentials{Username: "admin", Password: "wrongpassword"}
		body, _ := json.Marshal(creds)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		s.HandleLogin(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestHandleCounter(t *testing.T) {
	s := setupServer(t)

	// Helper to get token (logic moved to use HandleLogin)

	t.Run("GET should return counter value", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/counter", nil)
		rr := httptest.NewRecorder()

		s.HandleCounter(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp CounterResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Zero(t, resp.Value)
	})

	t.Run("POST should increment counter (authorized)", func(t *testing.T) {
		// We need to fetch/create a valid user first or generate token
		// Creating user
		_ = s.store.SaveUser("tester", "pass", "user")

		// Login to get real token
		creds := common.Credentials{Username: "tester", Password: "pass"}
		body, _ := json.Marshal(creds)
		lReq := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		lRr := httptest.NewRecorder()
		s.HandleLogin(lRr, lReq)
		var lResp map[string]string
		_ = json.Unmarshal(lRr.Body.Bytes(), &lResp)
		token := lResp["token"]

		req := httptest.NewRequest("POST", "/counter", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		s.HandleCounter(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp CounterResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal(t, uint64(1), resp.Value)
	})

	t.Run("DELETE should reset counter (admin only)", func(t *testing.T) {
		// Admin token
		creds := common.Credentials{Username: "admin", Password: "admin123"}
		body, _ := json.Marshal(creds)
		lReq := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		lRr := httptest.NewRecorder()
		s.HandleLogin(lRr, lReq)
		var lResp map[string]string
		_ = json.Unmarshal(lRr.Body.Bytes(), &lResp)
		adminToken := lResp["token"]

		req := httptest.NewRequest("DELETE", "/counter", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rr := httptest.NewRecorder()

		s.HandleCounter(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp CounterResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Zero(t, resp.Value)
	})

	t.Run("DELETE should fail for normal user", func(t *testing.T) {
		_ = s.store.SaveUser("normal", "pass", "user")
		// Get token
		creds := common.Credentials{Username: "normal", Password: "pass"}
		body, _ := json.Marshal(creds)
		lReq := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		lRr := httptest.NewRecorder()
		s.HandleLogin(lRr, lReq)
		var lResp map[string]string
		_ = json.Unmarshal(lRr.Body.Bytes(), &lResp)
		token := lResp["token"]

		req := httptest.NewRequest("DELETE", "/counter", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		s.HandleCounter(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

func TestHandleVersion(t *testing.T) {
	s := setupServer(t)

	req := httptest.NewRequest("GET", "/version", nil)
	rr := httptest.NewRecorder()

	s.HandleVersion(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp VersionResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "v1.2.3", resp.Version)
}
