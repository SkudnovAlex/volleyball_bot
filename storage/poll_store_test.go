package storage

import (
	"path/filepath"
	"testing"
)

func TestLoad_MissingFileReturnsEmpty(t *testing.T) {
	store := NewPollStore(filepath.Join(t.TempDir(), "polls.json"))

	polls, err := store.Load()
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(polls) != 0 {
		t.Fatalf("ожидался пустой список, получено: %d", len(polls))
	}
}

func TestAddLoadSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "polls.json")
	store := NewPollStore(path)

	if err := store.Add(PinnedPoll{MessageID: 1, CreatedAt: 1000}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := store.Add(PinnedPoll{MessageID: 2, CreatedAt: 2000}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	polls, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(polls) != 2 {
		t.Fatalf("ожидалось 2 записи, получено: %d", len(polls))
	}
	if polls[0].MessageID != 1 || polls[1].MessageID != 2 {
		t.Fatalf("неверные данные: %+v", polls)
	}

	// Save перезаписывает содержимое.
	if err := store.Save([]PinnedPoll{{MessageID: 9, CreatedAt: 9000}}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	polls, err = store.Load()
	if err != nil {
		t.Fatalf("Load после Save: %v", err)
	}
	if len(polls) != 1 || polls[0].MessageID != 9 {
		t.Fatalf("Save не перезаписал данные: %+v", polls)
	}
}

func TestLoad_PersistsAcrossInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "polls.json")

	first := NewPollStore(path)
	if err := first.Add(PinnedPoll{MessageID: 5, CreatedAt: 500}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	second := NewPollStore(path)
	polls, err := second.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(polls) != 1 || polls[0].MessageID != 5 {
		t.Fatalf("данные не сохранились на диск: %+v", polls)
	}
}
