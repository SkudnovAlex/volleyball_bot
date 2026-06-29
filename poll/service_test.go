package poll

import (
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"volleyball_bot/storage"
)

type mockSender struct {
	sent        []tgbotapi.Chattable
	requested   []tgbotapi.Chattable
	sendFunc    func(c tgbotapi.Chattable) (tgbotapi.Message, error)
	requestFunc func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

func (m *mockSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.sent = append(m.sent, c)
	if m.sendFunc != nil {
		return m.sendFunc(c)
	}
	return tgbotapi.Message{MessageID: 100}, nil
}

func (m *mockSender) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	m.requested = append(m.requested, c)
	if m.requestFunc != nil {
		return m.requestFunc(c)
	}
	return &tgbotapi.APIResponse{Ok: true}, nil
}

type mockRepo struct {
	polls   []storage.PinnedPoll
	loadErr error
	saveErr error
	saved   [][]storage.PinnedPoll
	added   []storage.PinnedPoll
}

func (r *mockRepo) Load() ([]storage.PinnedPoll, error) {
	return r.polls, r.loadErr
}

func (r *mockRepo) Save(polls []storage.PinnedPoll) error {
	r.saved = append(r.saved, polls)
	r.polls = polls
	return r.saveErr
}

func (r *mockRepo) Add(p storage.PinnedPoll) error {
	r.added = append(r.added, p)
	r.polls = append(r.polls, p)
	return nil
}

// testGameDays — независимый от приложения набор игровых дней для тестов.
var testGameDays = map[time.Weekday]bool{
	time.Tuesday:  true,
	time.Thursday: true,
	time.Friday:   true,
	time.Sunday:   true,
}

func tuesday() time.Time {
	// 30.06.2026 — вторник (есть в testGameDays).
	return time.Date(2026, time.June, 30, 20, 0, 0, 0, time.UTC)
}

func wednesday() time.Time {
	// 01.07.2026 — среда (нет в testGameDays).
	return time.Date(2026, time.July, 1, 20, 0, 0, 0, time.UTC)
}

func TestCheckAndCreatePoll_NonGameDay(t *testing.T) {
	sender := &mockSender{}
	repo := &mockRepo{}
	svc := New(sender, repo, 1, time.UTC, testGameDays)
	svc.now = wednesday

	svc.CheckAndCreatePoll()

	if len(sender.sent) != 0 {
		t.Fatalf("в нерабочий день опрос не должен отправляться, отправлено: %d", len(sender.sent))
	}
	if len(repo.added) != 0 {
		t.Fatalf("в нерабочий день ничего не должно сохраняться")
	}
}

func TestCheckAndCreatePoll_GameDay(t *testing.T) {
	sender := &mockSender{
		sendFunc: func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{MessageID: 555}, nil
		},
	}
	repo := &mockRepo{}
	svc := New(sender, repo, 42, time.UTC, testGameDays)
	svc.now = tuesday

	svc.CheckAndCreatePoll()

	if len(sender.sent) != 1 {
		t.Fatalf("ожидалась отправка одного опроса, получено: %d", len(sender.sent))
	}
	pollCfg, ok := sender.sent[0].(tgbotapi.SendPollConfig)
	if !ok {
		t.Fatalf("отправлен не SendPollConfig: %T", sender.sent[0])
	}
	if pollCfg.ChatID != 42 {
		t.Errorf("неверный chatID: %d", pollCfg.ChatID)
	}
	if len(pollCfg.Options) != len(pollOptions) {
		t.Errorf("неверное число вариантов: %d", len(pollCfg.Options))
	}

	// Закрепление было запрошено.
	if len(sender.requested) != 1 {
		t.Fatalf("ожидался один запрос (pin), получено: %d", len(sender.requested))
	}
	if _, ok := sender.requested[0].(tgbotapi.PinChatMessageConfig); !ok {
		t.Errorf("ожидался PinChatMessageConfig, получено: %T", sender.requested[0])
	}

	// Опрос сохранён с верным ID.
	if len(repo.added) != 1 || repo.added[0].MessageID != 555 {
		t.Fatalf("опрос должен быть сохранён с MessageID=555, added=%+v", repo.added)
	}
}

func TestDeleteOldPolls_RemovesOldKeepsFresh(t *testing.T) {
	now := tuesday()
	old := storage.PinnedPoll{MessageID: 1, CreatedAt: now.AddDate(0, 0, -daysAhead-1).Unix()}
	exactly := storage.PinnedPoll{MessageID: 2, CreatedAt: now.AddDate(0, 0, -daysAhead).Unix()}
	fresh := storage.PinnedPoll{MessageID: 3, CreatedAt: now.AddDate(0, 0, -1).Unix()}

	repo := &mockRepo{polls: []storage.PinnedPoll{old, exactly, fresh}}
	sender := &mockSender{}
	svc := New(sender, repo, 1, time.UTC, testGameDays)

	svc.deleteOldPolls(now)

	// Удалены old и exactly (>= порога), fresh остаётся.
	if len(sender.requested) != 2 {
		t.Fatalf("ожидалось 2 запроса на удаление, получено: %d", len(sender.requested))
	}
	if len(repo.saved) != 1 {
		t.Fatalf("ожидалось одно сохранение, получено: %d", len(repo.saved))
	}
	final := repo.saved[0]
	if len(final) != 1 || final[0].MessageID != 3 {
		t.Fatalf("в хранилище должен остаться только опрос 3, получено: %+v", final)
	}
}

func TestDeleteOldPolls_InsufficientRightsKeepsRecord(t *testing.T) {
	now := tuesday()
	old := storage.PinnedPoll{MessageID: 7, CreatedAt: now.AddDate(0, 0, -daysAhead-1).Unix()}

	repo := &mockRepo{polls: []storage.PinnedPoll{old}}
	sender := &mockSender{
		requestFunc: func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
			return nil, &tgbotapi.Error{Code: 400, Message: "Bad Request: not enough rights to delete a message"}
		},
	}
	svc := New(sender, repo, 1, time.UTC, testGameDays)

	svc.deleteOldPolls(now)

	final := repo.saved[0]
	if len(final) != 1 || final[0].MessageID != 7 {
		t.Fatalf("при нехватке прав запись должна сохраниться, получено: %+v", final)
	}
}

func TestDeleteOldPolls_EmptyStore(t *testing.T) {
	repo := &mockRepo{}
	sender := &mockSender{}
	svc := New(sender, repo, 1, time.UTC, testGameDays)

	svc.deleteOldPolls(tuesday())

	if len(sender.requested) != 0 {
		t.Fatalf("при пустом хранилище удалений быть не должно")
	}
	if len(repo.saved) != 0 {
		t.Fatalf("при пустом хранилище сохранений быть не должно")
	}
}
