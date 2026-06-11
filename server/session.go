package main

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	mathrand "math/rand"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/raaz/server/store"
)

// CrisisPauseDuration is how long message routing is suspended after a crisis
// trigger. Overridable in tests via package-level assignment.
var CrisisPauseDuration = 30 * time.Second

// Session holds shared state for a matched pair of clients.
type Session struct {
	id   string
	a, b *Client
	// paused is set to true when a crisis is detected; the relay goroutines
	// drop messages while paused and resume after CrisisPauseDuration.
	paused atomic.Bool
	ss     store.SessionStore
}

// triggerCrisis pauses the session, sends CRISIS_TRIGGERED to both participants,
// and schedules an auto-resume after CrisisPauseDuration. No-op if already paused.
func (sess *Session) triggerCrisis() {
	if sess.paused.Swap(true) {
		return // already in crisis mode
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
		time.Sleep(CrisisPauseDuration)
		sess.paused.Store(false)
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
	sess := &Session{id: generateID(), a: a, b: b, ss: ss}

	a.alias = randomAlias()
	b.alias = randomAlias()
	for b.alias == a.alias {
		b.alias = randomAlias()
	}

	slog.Info("session created", "sessionID", sess.id, "aliasA", a.alias, "aliasB", b.alias, "promptID", a.params.PromptID)

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
// When src.recv is closed (src disconnected), notifies dst and closes dst.conn.
func runRelay(src, dst *Client, sess *Session) {
	defer func() {
		notifyDisconnect(dst)
		// Allow writePump to flush the DISCONNECT message before forcing close.
		time.Sleep(150 * time.Millisecond)
		dst.conn.Close()
	}()

	for msg := range src.recv {
		if sess.paused.Load() {
			continue // crisis mode — drop message
		}
		out := routeMessage(msg, src, sess)
		if out == nil {
			continue // blocked by moderation or unknown frame
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
