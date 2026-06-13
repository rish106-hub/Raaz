package main

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	mathrand "math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/raaz/server/store"
)

// CrisisPauseDuration is how long message routing is suspended after a crisis trigger.
var CrisisPauseDuration = 30 * time.Second

// Session holds shared state for a matched pair of clients.
type Session struct {
	id         string
	a, b       *Client
	paused     atomic.Bool
	extensions atomic.Int32 // approved extensions granted; max 1
	done       chan struct{}
	closeOnce  sync.Once
	ss         store.SessionStore
}

func newSession(id string, a, b *Client, ss store.SessionStore) *Session {
	return &Session{id: id, a: a, b: b, done: make(chan struct{}), ss: ss}
}

// close signals that the session has ended, allowing goroutines blocked on done to exit.
func (sess *Session) close() {
	sess.closeOnce.Do(func() { close(sess.done) })
}

// triggerCrisis pauses the session, sends CRISIS_TRIGGERED to both participants,
// and schedules an auto-resume. No-op if already paused. The goroutine exits
// cleanly if the session ends before the pause timer fires (L-1 goroutine leak fix).
func (sess *Session) triggerCrisis() {
	if sess.paused.Swap(true) {
		return
	}
	data, _ := marshalEnvelope(EventCrisisTriggered, CrisisTriggeredPayload{
		Helplines: []string{
			"iCall: 9152987821",
			"Vandrevala Foundation: 1860-2662-345",
			"AASRA: 82-29-99-99-99",
			"Snehi: 044-24640050",
		},
		Message: "You don't have to face this alone. Please reach out to a helpline.",
	})
	sess.a.safeSend(data)
	sess.b.safeSend(data)
	go func() {
		select {
		case <-time.After(CrisisPauseDuration):
			sess.paused.Store(false)
		case <-sess.done:
			// session ended before pause timer; exit without leaking
		}
	}()
}

// generateID returns a UUID-v4-like hex string using crypto/rand.
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// randomAlias generates a server-assigned display name in the "Raaz #NNNN" format.
func randomAlias() string {
	return fmt.Sprintf("Raaz #%04d", mathrand.Intn(9000)+1000)
}

// createSession pairs two clients: assigns aliases, persists the session record,
// sends CONNECTED to both, then starts bidirectional relay goroutines.
func createSession(a, b *Client, ss store.SessionStore) {
	now := time.Now()
	sess := newSession(generateID(), a, b, ss)

	a.alias = randomAlias()
	b.alias = randomAlias()
	// L-4: bound alias collision loop to prevent infinite spin on degenerate RNG
	for attempt := 0; attempt < 10 && b.alias == a.alias; attempt++ {
		b.alias = randomAlias()
	}

	slog.Info("session created",
		"sessionID", sess.id,
		"aliasA", a.alias,
		"aliasB", b.alias,
		"promptID", a.params.PromptID,
	)

	rec := store.SessionRecord{
		SessionID: sess.id,
		AID:       a.params.AnonymousID,
		BID:       b.params.AnonymousID,
		AAlias:    a.alias,
		BAlias:    b.alias,
		PromptID:  a.params.PromptID,
		StartedAt: now,
		ExpiresAt: now.Add(time.Duration(SessionDuration) * time.Second),
	}
	if err := ss.Save(rec); err != nil {
		slog.Warn("save session error", "sessionID", sess.id, "err", err)
	}

	sendConnected(a, sess.id, b.alias)
	sendConnected(b, sess.id, a.alias)

	go runRelay(a, b, sess)
	go runRelay(b, a, sess)
}

// sendConnected emits the CONNECTED event to the given client.
func sendConnected(c *Client, matchID, partnerAlias string) {
	data, err := marshalEnvelope(EventConnected, ConnectedPayload{
		MatchID:                matchID,
		PartnerAlias:           partnerAlias,
		SessionDurationSeconds: SessionDuration,
	})
	if err != nil {
		slog.Error("marshal CONNECTED error", "err", err)
		return
	}
	c.safeSend(data)
}

// runRelay forwards messages from src.recv to dst.send through the moderation
// pipeline. Drops frames while sess.paused is true (crisis mode).
// On src disconnect, closes the session and notifies dst.
func runRelay(src, dst *Client, sess *Session) {
	defer func() {
		sess.close()
		notifyDisconnect(dst)
		time.Sleep(150 * time.Millisecond)
		dst.conn.Close()
	}()

	for msg := range src.recv {
		if sess.paused.Load() {
			continue
		}
		out := routeMessage(msg, src, sess)
		if out == nil {
			continue
		}
		dst.safeSend(out)
	}
}

// notifyDisconnect enqueues a DISCONNECT event in the partner's send channel.
func notifyDisconnect(c *Client) {
	data, err := marshalEnvelope(EventDisconnect, DisconnectPayload{Reason: "partner disconnected"})
	if err != nil {
		return
	}
	c.safeSend(data)
}

// disconnectMsg is a pre-serialised DISCONNECT frame for use in tests.
func disconnectMsg() []byte {
	data, _ := marshalEnvelope(EventDisconnect, DisconnectPayload{Reason: "partner disconnected"})
	return data
}

// isWebSocketCloseError reports whether err is a normal WebSocket close.
func isWebSocketCloseError(err error) bool {
	return websocket.IsCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
	)
}
