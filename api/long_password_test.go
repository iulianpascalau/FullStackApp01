package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"FullStackApp01/common"
	"FullStackApp01/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleChangePassword_LongPassword_Reproduction(t *testing.T) {
	s := mock.NewMockStorage()
	server := NewServer(s, testVersion, []byte("secret"))

	username := "longpassuser"
	// Register user with normal password first because SaveUser also uses bcrypt and would fail with long password
	err := s.SaveUser(username, "normalpass", "user")
	require.NoError(t, err)

	// Login to get token
	creds := common.Credentials{Username: username, Password: "normalpass"}
	body, _ := json.Marshal(creds)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	server.HandleLogin(rr, req)
	var resp map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	token := resp["token"]

	t.Run("should handle extremely long new password gracefully", func(t *testing.T) {
		longPass := strings.Repeat("a", 4096)
		reqData := ChangePasswordRequest{
			OldPassword: "normalpass",
			NewPassword: longPass,
		}
		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		// Helper to catch panic if any
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Server panicked: %v", r)
			}
		}()

		server.HandleChangePassword(rr, req)

		// Expecting 400 Bad Request
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		t.Logf("Response Code: %d", rr.Code)
		t.Logf("Response Body: %s", rr.Body.String())
	})

	t.Run("should handle extremely long old password gracefully", func(t *testing.T) {
		longPass := strings.Repeat("b", 8192)
		reqData := ChangePasswordRequest{
			OldPassword: longPass,
			NewPassword: "short",
		}
		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		server.HandleChangePassword(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		t.Logf("Response Code: %d", rr.Code)
		t.Logf("Response Body: %s", rr.Body.String())
	})
}
