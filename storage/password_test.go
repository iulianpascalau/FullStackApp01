package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestStore_UpdatePassword(t *testing.T) {
	s, err := NewStore(t.TempDir())
	require.NoError(t, err)
	defer func() {
		_ = s.Close()
	}()

	username := "updater"
	password := "oldpassword"
	err = s.SaveUser(username, password, "user")
	require.NoError(t, err)

	t.Run("should update password successfully", func(t *testing.T) {
		newPassword := "newpassword"
		err := s.UpdatePassword(username, newPassword)
		assert.NoError(t, err)

		user, err := s.GetUser(username)
		assert.NoError(t, err)
		assert.NoError(t, bcrypt.CompareHashAndPassword(user.Hash, []byte(newPassword)))
		assert.Error(t, bcrypt.CompareHashAndPassword(user.Hash, []byte(password)))
	})

	t.Run("should fail for non-existent user", func(t *testing.T) {
		err := s.UpdatePassword("ghost", "pass")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}
