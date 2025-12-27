package mock

import (
	"errors"

	"FullStackApp01/common"

	"golang.org/x/crypto/bcrypt"
)

type mockStorage struct {
	counter uint64
	users   map[string]*common.User
}

// NewMockStorage -
func NewMockStorage() *mockStorage {
	return &mockStorage{
		users: make(map[string]*common.User),
	}
}

// GetUser -
func (mock *mockStorage) GetUser(username string) (*common.User, error) {
	data, ok := mock.users[username]
	if !ok {
		return nil, errors.New("user not found")
	}

	return data, nil
}

// SaveUser -
func (mock *mockStorage) SaveUser(username, password string, role string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	mock.users[username] = &common.User{
		Username: username,
		Role:     role,
		Hash:     hash,
	}

	return nil
}

// ResetCounter -
func (mock *mockStorage) ResetCounter() error {
	mock.counter = 0

	return nil
}

// IncrementCounter -
func (mock *mockStorage) IncrementCounter() (uint64, error) {
	mock.counter++
	return mock.counter, nil
}

// GetCounter -
func (mock *mockStorage) GetCounter() (uint64, error) {
	return mock.counter, nil
}

func (mock *mockStorage) Close() error {
	return nil
}

// UpdatePassword -
func (mock *mockStorage) UpdatePassword(username, newPassword string) error {
	data, ok := mock.users[username]
	if !ok {
		return errors.New("user not found")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	data.Hash = hash
	return nil
}
