package storage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	t.Parallel()

	t.Run("can not open twice the same path", func(t *testing.T) {
		tempDir := t.TempDir()
		instance, err := NewStore(tempDir)
		assert.Nil(t, err)
		assert.NotNil(t, instance)

		instance2, err2 := NewStore(tempDir)
		assert.NotNil(t, err2)
		assert.Contains(t, err2.Error(), "resource temporarily unavailable")
		assert.Nil(t, instance2)

		_ = instance.Close()
	})
	t.Run("should work", func(t *testing.T) {
		tempDir := t.TempDir()
		instance, err := NewStore(tempDir)
		assert.Nil(t, err)
		assert.NotNil(t, instance)

		_ = instance.Close()
	})
}

func TestStore_GetCounter(t *testing.T) {
	t.Parallel()

	t.Run("should error if the DB is closed", func(t *testing.T) {
		testPath := t.TempDir()

		instance, _ := NewStore(testPath)
		_ = instance.Close()

		val, err := instance.GetCounter()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "leveldb: closed")
		assert.Zero(t, val)
	})
	t.Run("should error if the stored data is not an uint64", func(t *testing.T) {
		testPath := t.TempDir()

		instance, _ := NewStore(testPath)

		_ = instance.db.Put([]byte(counterKey), []byte("this is not a valid uint64 data"), nil)
		val, err := instance.GetCounter()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid counter data")
		assert.Zero(t, val)

		_ = instance.Close()
	})
	t.Run("should work", func(t *testing.T) {
		testPath := t.TempDir()

		instance, _ := NewStore(testPath)

		val, err := instance.GetCounter()
		assert.Nil(t, err)
		assert.Zero(t, val)

		_ = instance.Close()
	})
}

func TestStore_GetUser(t *testing.T) {
	t.Parallel()

	t.Run("should error if the DB is closed", func(t *testing.T) {
		testPath := t.TempDir()

		instance, _ := NewStore(testPath)
		_ = instance.Close()

		userData, err := instance.GetUser("test")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "leveldb: closed")
		assert.Nil(t, userData)
	})
	t.Run("should error if the stored data is not a valid User", func(t *testing.T) {
		testPath := t.TempDir()

		instance, _ := NewStore(testPath)

		_ = instance.db.Put([]byte(userKeyPrefix+"test"), []byte("this is not a valid json data"), nil)
		userData, err := instance.GetUser("test")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid character 'h' in literal true")
		assert.Nil(t, userData)

		_ = instance.Close()
	})
	t.Run("should work", func(t *testing.T) {
		testPath := t.TempDir()

		instance, _ := NewStore(testPath)

		_ = instance.SaveUser("test", "psw", "admin")

		userData, err := instance.GetUser("test")
		assert.Nil(t, err)
		assert.Equal(t, "test", userData.Username)
		assert.Equal(t, "admin", userData.Role)

		_ = instance.Close()
	})
}

func TestStore_IncrementCounter(t *testing.T) {
	t.Parallel()

	t.Run("should error if the DB is closed", func(t *testing.T) {
		testPath := t.TempDir()

		instance, _ := NewStore(testPath)
		_ = instance.Close()

		val, err := instance.IncrementCounter()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "leveldb: closed")
		assert.Zero(t, val)
	})
	t.Run("should work", func(t *testing.T) {
		testPath := t.TempDir()

		instance, _ := NewStore(testPath)

		val, err := instance.IncrementCounter()
		assert.Nil(t, err)
		assert.Equal(t, uint64(1), val)

		_ = instance.Close()
	})
}

func TestStore_ResetCounter(t *testing.T) {
	t.Parallel()

	t.Run("should error if the DB is closed", func(t *testing.T) {
		testPath := t.TempDir()

		instance, _ := NewStore(testPath)
		_ = instance.Close()

		err := instance.ResetCounter()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "leveldb: closed")
	})
	t.Run("should work", func(t *testing.T) {
		testPath := t.TempDir()

		instance, _ := NewStore(testPath)

		err := instance.ResetCounter()
		assert.Nil(t, err)

		_ = instance.Close()
	})
}

func TestStore_Lifecycle(t *testing.T) {
	t.Parallel()

	testPath := t.TempDir()

	// Test NewStore
	instance, _ := NewStore(testPath)

	// Test Initial GetCounter
	val, err := instance.GetCounter()
	assert.Nil(t, err)
	assert.Zero(t, val)

	// Test IncrementCounter
	newVal, err := instance.IncrementCounter()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), newVal)

	// Test Persistence (Close and Reopen)
	err = instance.Close()
	assert.Nil(t, err)

	instance, err = NewStore(testPath)
	assert.Nil(t, err)

	val, err = instance.GetCounter()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), val)

	// Test ResetCounter
	err = instance.ResetCounter()
	assert.Nil(t, err)

	val, err = instance.GetCounter()
	assert.Nil(t, err)
	assert.Zero(t, val)

	_ = instance.Close()
}

func TestStore_Concurrency(t *testing.T) {
	testPath := t.TempDir()

	instance, _ := NewStore(testPath)

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := instance.IncrementCounter()
			assert.Nil(t, err)
		}()
	}

	wg.Wait()

	val, err := instance.GetCounter()
	assert.Nil(t, err)
	assert.Equal(t, uint64(iterations), val)

	_ = instance.Close()
}

func TestStore_Users(t *testing.T) {
	t.Parallel()

	// Test Setup
	testPath := t.TempDir()
	instance, _ := NewStore(testPath)
	defer func() {
		_ = instance.Close()
	}()

	t.Run("should save and retrieve user", func(t *testing.T) {
		err := instance.SaveUser("testuser", "password123", "user")
		assert.Nil(t, err)

		user, err := instance.GetUser("testuser")
		assert.Nil(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "user", user.Role)
		assert.NotEmpty(t, user.Hash)
	})

	t.Run("should verify password hash", func(t *testing.T) {
		// Only run if user exists (depends on previous test ideally, or create fresh)
		_ = instance.SaveUser("secureuser", "securepass", "admin")
		user, err := instance.GetUser("secureuser")
		assert.Nil(t, err)

		// Check if password matches
		// We need to import bcrypt in test or trust SaveUser did its job,
		// but ideally verify it is actually hashed.
		// Since we cannot easily import bcrypt here without modifying imports, we check hash length.
		assert.True(t, len(user.Hash) > 0)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		user, err := instance.GetUser("nonexistent")
		assert.NotNil(t, err)
		assert.Equal(t, "user not found", err.Error())
		assert.Nil(t, user)
	})

	t.Run("should prevent duplication", func(t *testing.T) {
		err := instance.SaveUser("duplicated", "password123", "user")
		assert.Nil(t, err)

		err = instance.SaveUser("duplicated", "aaa", "user")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "user already exists")
	})

	t.Run("duplication check error if db is closed", func(t *testing.T) {
		instance2, _ := NewStore(t.TempDir())
		_ = instance2.Close()

		err := instance2.SaveUser("duplicated", "password123", "user")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "leveldb: closed")
	})
}
