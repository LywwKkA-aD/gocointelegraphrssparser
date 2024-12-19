// internal/repository/users.go
package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/LywwKkA-aD/gocointelegraphrssparser/pkg/logger"
)

type UserRepository struct {
	users    map[int64]bool
	filePath string
	mu       sync.RWMutex
}

func NewUserRepository(dataDir string) (*UserRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dataDir, "users.json")
	repo := &UserRepository{
		users:    make(map[int64]bool),
		filePath: filePath,
	}

	if err := repo.load(); err != nil {
		logger.Warn("No existing users file found or error loading: %v", err)
	}

	return repo, nil
}

func (r *UserRepository) Add(userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[userID] = true
	return r.save()
}

func (r *UserRepository) Remove(userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.users, userID)
	return r.save()
}

func (r *UserRepository) GetAll() map[int64]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy to prevent external modifications
	users := make(map[int64]bool, len(r.users))
	for k, v := range r.users {
		users[k] = v
	}
	return users
}

func (r *UserRepository) save() error {
	data, err := json.Marshal(r.users)
	if err != nil {
		return err
	}
	return os.WriteFile(r.filePath, data, 0644)
}

func (r *UserRepository) load() error {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &r.users)
}
