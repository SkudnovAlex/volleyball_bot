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

func countType[T tgbotapi.Chattable](reqs []tgbotapi.Chattable) int {
	n := 0
	for _, r := range reqs {
		if _, ok := r.(T); ok {
			n++
		}
	}
	return n
}

func TestCreatePoll_NonGameDay(t *testing.T) {
	sender := &mockSender{}
	repo := &mockRepo{}
	svc := New(sender, repo, 1, time.UTC, testGameDays)
	svc.now = wednesday

	svc.CreatePoll()

	if len(sender.sent) != 0 {
		t.Fatalf("в нерабочий день опрос не должен отправляться, отправлено: %d", len(sender.sent))
	}
	if len(repo.added) != 0 {
		t.Fatalf("в нерабочий день ничего не должно сохраняться")
	}
}

func TestCreatePoll_GameDay(t *testing.T) {
	sender := &mockSender{
		sendFunc: func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{MessageID: 555}, nil
		},
	}
	repo := &mockRepo{}
	svc := New(sender, repo, 42, time.UTC, testGameDays)
	svc.now = tuesday

	svc.CreatePoll()

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

	// Закрепление было запрошено, удаление — нет.
	if countType[tgbotapi.PinChatMessageConfig](sender.requested) != 1 {
		t.Fatalf("ожидался один запрос на закрепление, requested=%+v", sender.requested)
	}
	if countType[tgbotapi.DeleteMessageConfig](sender.requested) != 0 {
		t.Fatalf("создание не должно ничего удалять")
	}

	// Опрос сохранён с верным ID, датой игры и флагом Pinned.
	if len(repo.added) != 1 || repo.added[0].MessageID != 555 || !repo.added[0].Pinned {
		t.Fatalf("опрос должен быть сохранён с MessageID=555 и Pinned=true, added=%+v", repo.added)
	}
	if repo.added[0].GameDate != gameDate(tuesday()).Unix() {
		t.Errorf("GameDate = %d, ожидалось %d", repo.added[0].GameDate, gameDate(tuesday()).Unix())
	}
}

func TestCreatePoll_SkipIfAlreadyExists(t *testing.T) {
	existing := storage.PinnedPoll{MessageID: 1, CreatedAt: tuesday().Unix(), GameDate: gameDate(tuesday()).Unix(), Pinned: true}
	repo := &mockRepo{polls: []storage.PinnedPoll{existing}}
	sender := &mockSender{}
	svc := New(sender, repo, 1, time.UTC, testGameDays)
	svc.now = tuesday

	svc.CreatePoll()

	if len(sender.sent) != 0 {
		t.Fatalf("опрос на эту дату уже есть — создавать не нужно, отправлено: %d", len(sender.sent))
	}
	if len(repo.added) != 0 {
		t.Fatalf("дубль не должен сохраняться")
	}
}

func TestCleanupPolls_DeletesOldKeepsToday(t *testing.T) {
	now := tuesday()
	old := storage.PinnedPoll{MessageID: 1, CreatedAt: now.AddDate(0, 0, -deleteAfterDays-1).Unix(), Pinned: true}
	exactly := storage.PinnedPoll{MessageID: 2, CreatedAt: now.AddDate(0, 0, -deleteAfterDays).Unix(), Pinned: true}
	today := storage.PinnedPoll{MessageID: 3, CreatedAt: now.Unix(), Pinned: true}

	repo := &mockRepo{polls: []storage.PinnedPoll{old, exactly, today}}
	sender := &mockSender{}
	svc := New(sender, repo, 1, time.UTC, testGameDays)
	svc.now = tuesday

	svc.CleanupPolls()

	// Удалены old и exactly (возраст >= 7 дней), сегодняшний остаётся закреплённым.
	if got := countType[tgbotapi.DeleteMessageConfig](sender.requested); got != 2 {
		t.Fatalf("ожидалось 2 удаления, получено: %d", got)
	}
	if got := countType[tgbotapi.UnpinChatMessageConfig](sender.requested); got != 0 {
		t.Fatalf("сегодняшний опрос не должен открепляться, откреплений: %d", got)
	}
	final := repo.saved[0]
	if len(final) != 1 || final[0].MessageID != 3 || !final[0].Pinned {
		t.Fatalf("должен остаться только опрос 3 (Pinned), получено: %+v", final)
	}
}

func TestCleanupPolls_UnpinsYesterday(t *testing.T) {
	now := tuesday()
	yesterday := storage.PinnedPoll{MessageID: 9, CreatedAt: now.AddDate(0, 0, -1).Unix(), Pinned: true}

	repo := &mockRepo{polls: []storage.PinnedPoll{yesterday}}
	sender := &mockSender{}
	svc := New(sender, repo, 1, time.UTC, testGameDays)
	svc.now = tuesday

	svc.CleanupPolls()

	if got := countType[tgbotapi.UnpinChatMessageConfig](sender.requested); got != 1 {
		t.Fatalf("ожидалось 1 открепление, получено: %d", got)
	}
	if got := countType[tgbotapi.DeleteMessageConfig](sender.requested); got != 0 {
		t.Fatalf("вчерашний опрос не должен удаляться, удалений: %d", got)
	}
	final := repo.saved[0]
	if len(final) != 1 || final[0].MessageID != 9 || final[0].Pinned {
		t.Fatalf("опрос 9 должен остаться с Pinned=false, получено: %+v", final)
	}
}

func TestCleanupPolls_KeepsTodayPinned(t *testing.T) {
	now := tuesday()
	today := storage.PinnedPoll{MessageID: 5, CreatedAt: now.Unix(), Pinned: true}

	repo := &mockRepo{polls: []storage.PinnedPoll{today}}
	sender := &mockSender{}
	svc := New(sender, repo, 1, time.UTC, testGameDays)
	svc.now = tuesday

	svc.CleanupPolls()

	if len(sender.requested) != 0 {
		t.Fatalf("сегодняшний опрос трогать не нужно, requested=%+v", sender.requested)
	}
	final := repo.saved[0]
	if len(final) != 1 || !final[0].Pinned {
		t.Fatalf("сегодняшний опрос должен остаться закреплённым, получено: %+v", final)
	}
}

func TestCleanupPolls_AlreadyUnpinnedSkipsUnpin(t *testing.T) {
	now := tuesday()
	yesterday := storage.PinnedPoll{MessageID: 8, CreatedAt: now.AddDate(0, 0, -1).Unix(), Pinned: false}

	repo := &mockRepo{polls: []storage.PinnedPoll{yesterday}}
	sender := &mockSender{}
	svc := New(sender, repo, 1, time.UTC, testGameDays)
	svc.now = tuesday

	svc.CleanupPolls()

	if len(sender.requested) != 0 {
		t.Fatalf("уже откреплённый опрос трогать не нужно, requested=%+v", sender.requested)
	}
}

func TestCleanupPolls_InsufficientRightsKeepsRecord(t *testing.T) {
	now := tuesday()
	old := storage.PinnedPoll{MessageID: 7, CreatedAt: now.AddDate(0, 0, -deleteAfterDays-1).Unix(), Pinned: true}

	repo := &mockRepo{polls: []storage.PinnedPoll{old}}
	sender := &mockSender{
		requestFunc: func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
			return nil, &tgbotapi.Error{Code: 400, Message: "Bad Request: not enough rights to delete a message"}
		},
	}
	svc := New(sender, repo, 1, time.UTC, testGameDays)
	svc.now = tuesday

	svc.CleanupPolls()

	final := repo.saved[0]
	if len(final) != 1 || final[0].MessageID != 7 {
		t.Fatalf("при нехватке прав запись должна сохраниться, получено: %+v", final)
	}
}

func TestCleanupPolls_EmptyStore(t *testing.T) {
	repo := &mockRepo{}
	sender := &mockSender{}
	svc := New(sender, repo, 1, time.UTC, testGameDays)
	svc.now = tuesday

	svc.CleanupPolls()

	if len(sender.requested) != 0 {
		t.Fatalf("при пустом хранилище запросов быть не должно")
	}
	if len(repo.saved) != 0 {
		t.Fatalf("при пустом хранилище сохранений быть не должно")
	}
}
