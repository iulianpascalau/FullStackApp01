package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"FullStackApp01/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORSHeaders_OnError(t *testing.T) {
	s := mock.NewMockStorage()
	server := NewServer(s, testVersion, []byte("secret"))

	// Create a user
	username := "corsuser"
	_ = s.SaveUser(username, "pass", "user")

	t.Run("should output CORS headers even on 400 error", func(t *testing.T) {
		// Send request with long password to trigger 400
		longPass := strings.Repeat("a", 100)
		reqData := ChangePasswordRequest{
			OldPassword: "pass",
			NewPassword: longPass,
		}
		body, _ := json.Marshal(reqData)

		// Authenticate
		// Login shim
		// ... actually HandleChangePassword needs valid token to reach validation
		// Lets mock a request with valid token header
		// We need a helper to generate token

		// Shortcut: Bypass auth check? No, headers set before auth.
		// Sending garbage auth header -> 401 Unauthorized from Authorized?
		// No, GetUserFromToken returns error.
		// HandleChangePassword returns 401.

		req := httptest.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer invalidtoken")
		rr := httptest.NewRecorder()

		server.HandleChangePassword(rr, req)

		// Should receive 401 or 500 or 400 depending on where it stops
		// But headers MUST be there
		assert.NotEmpty(t, rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should output CORS headers on 400 validation error", func(t *testing.T) {
		// Valid token, invalid body
		// We need valid token to reach 400 check
		// Login first
		lBody := `{"username":"corsuser","password":"pass"}`
		lReq := httptest.NewRequest("POST", "/login", bytes.NewBufferString(lBody))
		lRr := httptest.NewRecorder()
		server.HandleLogin(lRr, lReq)
		var lResp map[string]string
		_ = json.Unmarshal(lRr.Body.Bytes(), &lResp)
		token := lResp["token"]
		require.NotEmpty(t, token)

		longPass := strings.Repeat("a", 100)
		reqData := ChangePasswordRequest{
			OldPassword: "pass",
			NewPassword: longPass,
		}
		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		server.HandleChangePassword(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	})
}
