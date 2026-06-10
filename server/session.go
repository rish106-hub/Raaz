package main

import (
	"crypto/rand"
	"fmt"
	"log"
	mathrand "math/rand"
	"time"

	"github.com/gorilla/websocket"
)

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

// createSession pairs two clients: assigns aliases, sends CONNECTED to both,
// then starts bidirectional relay goroutines.
func createSession(a, b *Client) {
	sessionID := generateID()

	a.alias = randomAlias()
	b.alias = randomAlias()
	for b.alias == a.alias {
		b.alias = randomAlias()
	}

	log.Printf("session %s: %q <-> %q (prompt=%s)", sessionID, a.alias, b.alias, a.params.PromptID)

	sendConnected(a, sessionID, b.alias)
	sendConnected(b, sessionID, a.alias)

	// Each relay goroutine reads from src.recv and writes to dst.send.
	// When src disconnects (recv closed), it notifies dst and closes dst.conn.
	go runRelay(a, b)
	go runRelay(b, a)
}

// sendConnected emits the CONNECTED event to the given client.
func sendConnected(c *Client, matchID, partnerAlias string) {
	data, err := marshalEnvelope(EventConnected, ConnectedPayload{
		MatchID:                matchID,
		PartnerAlias:           partnerAlias,
		SessionDurationSeconds: SessionDuration,
	})
	if err != nil {
		log.Printf("marshal CONNECTED: %v", err)
		return
	}
	c.safeSend(data)
}

// runRelay forwards all messages from src.recv to dst.send, rewriting
// senderAlias with the server-assigned alias to prevent identity spoofing.
// When src.recv is closed (src disconnected), it notifies dst and closes
// dst.conn so dst's pumps also exit cleanly.
func runRelay(src, dst *Client) {
	defer func() {
		notifyDisconnect(dst)
		// Allow writePump to flush the DISCONNECT message before forcing close.
		time.Sleep(150 * time.Millisecond)
		dst.conn.Close()
	}()

	for msg := range src.recv {
		out := routeMessage(msg, src)
		if out == nil {
			continue // invalid or unknown frame — dropped
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
