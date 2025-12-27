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

func TestHandleChangePassword(t *testing.T) {
	setupWithUser := func(t *testing.T) (*Server, string) {
		t.Helper()
		store := mock.NewMockStorage()

		username := "changer"
		password := "oldpass"
		err := store.SaveUser(username, password, "user")
		require.NoError(t, err)

		return NewServer(store, testKey), username
	}

	t.Run("should change password successfully", func(t *testing.T) {
		s, username := setupWithUser(t)

		// Since createTestToken isn't available, lets login
		creds := common.Credentials{Username: username, Password: "oldpass"}
		lBody, _ := json.Marshal(creds)
		lReq := httptest.NewRequest("POST", "/login", bytes.NewBuffer(lBody))
		lRr := httptest.NewRecorder()
		s.HandleLogin(lRr, lReq)
		var lResp map[string]string
		_ = json.Unmarshal(lRr.Body.Bytes(), &lResp)
		tokenStr := lResp["token"]

		// Change password
		reqData := ChangePasswordRequest{
			OldPassword: "oldpass",
			NewPassword: "newpass",
		}
		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		rr := httptest.NewRecorder()

		s.HandleChangePassword(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify new password works
		credsNew := common.Credentials{Username: username, Password: "newpass"}
		lBodyNew, _ := json.Marshal(credsNew)
		lReqNew := httptest.NewRequest("POST", "/login", bytes.NewBuffer(lBodyNew))
		lRrNew := httptest.NewRecorder()
		s.HandleLogin(lRrNew, lReqNew)
		assert.Equal(t, http.StatusOK, lRrNew.Code)
	})

	t.Run("should fail with wrong old password", func(t *testing.T) {
		s, username := setupWithUser(t)

		// Login
		creds := common.Credentials{Username: username, Password: "oldpass"}
		lBody, _ := json.Marshal(creds)
		lReq := httptest.NewRequest("POST", "/login", bytes.NewBuffer(lBody))
		lRr := httptest.NewRecorder()
		s.HandleLogin(lRr, lReq)
		var lResp map[string]string
		_ = json.Unmarshal(lRr.Body.Bytes(), &lResp)
		tokenStr := lResp["token"]

		reqData := ChangePasswordRequest{
			OldPassword: "wrongpass",
			NewPassword: "newpass",
		}
		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		rr := httptest.NewRecorder()

		s.HandleChangePassword(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
