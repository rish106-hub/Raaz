package store

import "time"

type StrikeAction int

const (
	StrikeActionWarn       StrikeAction = iota
	StrikeActionDisconnect
)

type StrikeResult struct {
	Strikes int
	Action  StrikeAction
}

// WaitingEntry is the serialisable form of a queued client.
type WaitingEntry struct {
	AnonymousID string
	PromptID    string
	AgeBucket   string
	City        string
	JoinedAt    time.Time
}

// SessionRecord holds active session metadata.
type SessionRecord struct {
	SessionID string
	AID       string
	BID       string
	AAlias    string
	BAlias    string
	PromptID  string
	StartedAt time.Time
	ExpiresAt time.Time
}

// VaultMessage is a server-side vault entry.
type VaultMessage struct {
	AnonIDHash  string
	MessageID   string
	SessionID   string
	Prompt      string
	Text        string
	SenderAlias string
	SavedAt     time.Time
}

type StrikeStore interface {
	RecordStrike(userID string) (StrikeResult, error)
	IsBanned(userID string) (bool, error)
}

type QueueStore interface {
	Enqueue(entry WaitingEntry) error
	Remove(anonymousID string) error
	Snapshot() ([]WaitingEntry, error)
	MigrateToFallback(olderThan time.Time) error
}

type SessionStore interface {
	Save(sess SessionRecord) error
	Get(sessionID string) (*SessionRecord, error)
	Delete(sessionID string) error
}

type VaultStore interface {
	SaveMessage(msg VaultMessage) error
	GetMessages(anonIDHash string) ([]VaultMessage, error)
}
