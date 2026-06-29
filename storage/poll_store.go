package storage

import (
	"encoding/json"
	"errors"
	"os"
)

func NewPollStore(path string) *PollStore {
	return &PollStore{path: path}
}

// Load читает список опросов. Если файла нет — возвращает пустой список.
func (s *PollStore) Load() ([]PinnedPoll, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *PollStore) loadLocked() ([]PinnedPoll, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []PinnedPoll{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []PinnedPoll{}, nil
	}
	var polls []PinnedPoll
	if err := json.Unmarshal(data, &polls); err != nil {
		return nil, err
	}
	return polls, nil
}

// Save атомарно перезаписывает файл хранилища.
func (s *PollStore) Save(polls []PinnedPoll) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(polls)
}

func (s *PollStore) saveLocked(polls []PinnedPoll) error {
	data, err := json.MarshalIndent(polls, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Add добавляет новый опрос в хранилище.
func (s *PollStore) Add(p PinnedPoll) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	polls, err := s.loadLocked()
	if err != nil {
		return err
	}
	polls = append(polls, p)
	return s.saveLocked(polls)
}
