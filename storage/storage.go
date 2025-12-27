package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"sync"

	"FullStackApp01/common"

	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/crypto/bcrypt"
)

const counterKey = "counter"
const userKeyPrefix = "user:"

// Store handles the persistence layer using LevelDB
type store struct {
	db *leveldb.DB
	mu sync.Mutex
}

// NewStore creates or opens a database at the given path
func NewStore(path string) (*store, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &store{db: db}, nil
}

// Close closes the underlying database
func (s *store) Close() error {
	return s.db.Close()
}

// GetCounter retrieves the current value of the counter
func (s *store) GetCounter() (uint64, error) {
	data, err := s.db.Get([]byte(counterKey), nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if len(data) != 8 {
		return 0, errors.New("invalid counter data")
	}
	return binary.BigEndian.Uint64(data), nil
}

// IncrementCounter increments the counter safely and returns the new value
func (s *store) IncrementCounter() (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, err := s.GetCounter()
	if err != nil {
		return 0, err
	}

	val++
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, val)

	err = s.db.Put([]byte(counterKey), buf, nil)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ErrUserAlreadyExists is returned when trying to create a user that already exists
var ErrUserAlreadyExists = errors.New("user already exists")

// SaveUser creates or updates a user with a hashed password
func (s *store) SaveUser(username, password, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists
	exists, err := s.db.Has([]byte(userKeyPrefix+username), nil)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := common.User{
		Username: username,
		Role:     role,
		Hash:     hash,
	}

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.db.Put([]byte(userKeyPrefix+username), data, nil)
}

// GetUser retrieves a user by username
func (s *store) GetUser(username string) (*common.User, error) {
	data, err := s.db.Get([]byte(userKeyPrefix+username), nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	var user common.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePassword updates the password for an existing user
func (s *store) UpdatePassword(username, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := []byte(userKeyPrefix + username)
	data, err := s.db.Get(key, nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		return errors.New("user not found")
	}
	if err != nil {
		return err
	}

	var user common.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Hash = hash

	newData, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.db.Put(key, newData, nil)
}

// ResetCounter resets the counter to 0
func (s *store) ResetCounter() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, 0)

	return s.db.Put([]byte(counterKey), buf, nil)
}
