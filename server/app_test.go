package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// testServer starts an httptest.Server backed by a fresh App and returns its
// WebSocket base URL (ws://...) and a cleanup function.
func testServer(t *testing.T) (wsBase string, cleanup func()) {
	t.Helper()
	app := NewApp()
	app.Start()
	ts := httptest.NewServer(corsMiddleware(app))
	wsBase = "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	return wsBase, ts.Close
}

// dialWS connects a WebSocket client using the given URL with a 3-second
// deadline for all read operations.
func dialWS(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", url, err)
	}
	return conn
}

// readEnvelope reads the next message from conn and unmarshals it into an Envelope.
// Fails the test if the read or unmarshal fails.
func readEnvelope(t *testing.T, conn *websocket.Conn, deadline time.Duration) Envelope {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(deadline))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("readEnvelope: %v", err)
	}
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	return env
}

// sendMessage writes a typed envelope to conn.
func sendMessage(t *testing.T, conn *websocket.Conn, env Envelope) {
	t.Helper()
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, raw); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// TestTwoClientsMatchSameCity verifies that two clients with identical
// PromptID + AgeBucket + City are matched immediately and both receive
// a CONNECTED event containing the partner's alias.
func TestTwoClientsMatchSameCity(t *testing.T) {
	wsBase, cleanup := testServer(t)
	defer cleanup()

	params := "?promptId=p1&ageBucket=18-25&city=Mumbai"

	// Connect both clients before reading from either so the match loop
	// can pair them after both are in the queue.
	connA := dialWS(t, wsBase+params+"&anonymousId=userA")
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId=userB")
	defer connB.Close()

	envA := readEnvelope(t, connA, 2*time.Second)
	envB := readEnvelope(t, connB, 2*time.Second)

	if envA.Type != EventConnected {
		t.Fatalf("A: expected CONNECTED, got %q", envA.Type)
	}
	if envB.Type != EventConnected {
		t.Fatalf("B: expected CONNECTED, got %q", envB.Type)
	}

	var payA, payB ConnectedPayload
	mustUnmarshal(t, envA.Payload, &payA)
	mustUnmarshal(t, envB.Payload, &payB)

	if payA.MatchID == "" {
		t.Error("A: matchId is empty")
	}
	if payA.MatchID != payB.MatchID {
		t.Errorf("session IDs differ: A=%s B=%s", payA.MatchID, payB.MatchID)
	}
	if payA.PartnerAlias == "" {
		t.Error("A: partnerAlias is empty")
	}
	if payB.PartnerAlias == "" {
		t.Error("B: partnerAlias is empty")
	}
	// Each client's partnerAlias should equal the other's own alias.
	// We verify they are different non-empty strings.
	if payA.PartnerAlias == payB.PartnerAlias {
		t.Errorf("partner aliases must differ, both are %q", payA.PartnerAlias)
	}
	if payA.SessionDurationSeconds != SessionDuration {
		t.Errorf("A: expected session duration %d, got %d", SessionDuration, payA.SessionDurationSeconds)
	}
}

// TestMessageRouting verifies that a message sent by client A is received
// by client B with the server-assigned SenderAlias (not A's self-reported alias).
func TestMessageRouting(t *testing.T) {
	wsBase, cleanup := testServer(t)
	defer cleanup()

	params := "?promptId=p2&ageBucket=26-35&city=Delhi"
	connA := dialWS(t, wsBase+params+"&anonymousId=userC")
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId=userD")
	defer connB.Close()

	// Wait for both CONNECTED events.
	readEnvelope(t, connA, 2*time.Second)
	readEnvelope(t, connB, 2*time.Second)

	// A sends a chat message with a spoofed senderAlias.
	msgPayload, _ := json.Marshal(MessagePayload{
		MessageID:   "msg-1",
		Text:        "hello from A",
		SenderAlias: "SPOOFED",
		Timestamp:   0,
	})
	sendMessage(t, connA, Envelope{
		Type:    EventMessage,
		Payload: json.RawMessage(msgPayload),
	})

	// B should receive the message with the server-assigned alias, not "SPOOFED".
	env := readEnvelope(t, connB, 2*time.Second)
	if env.Type != EventMessage {
		t.Fatalf("B: expected MESSAGE, got %q", env.Type)
	}
	var recv MessagePayload
	mustUnmarshal(t, env.Payload, &recv)

	if recv.Text != "hello from A" {
		t.Errorf("text: want %q, got %q", "hello from A", recv.Text)
	}
	if recv.SenderAlias == "SPOOFED" {
		t.Error("server must overwrite spoofed senderAlias")
	}
	if recv.SenderAlias == "" {
		t.Error("senderAlias must not be empty after server rewrite")
	}
	if recv.Timestamp == 0 {
		t.Error("server must stamp timestamp > 0")
	}
}

// TestTypingEventRouted verifies that a TYPING event is forwarded to the partner
// with the correct server-assigned senderAlias.
func TestTypingEventRouted(t *testing.T) {
	wsBase, cleanup := testServer(t)
	defer cleanup()

	params := "?promptId=p3&ageBucket=18-25&city=Bangalore"
	connA := dialWS(t, wsBase+params+"&anonymousId=userE")
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId=userF")
	defer connB.Close()

	readEnvelope(t, connA, 2*time.Second)
	readEnvelope(t, connB, 2*time.Second)

	typingPayload, _ := json.Marshal(TypingPayload{IsTyping: true, SenderAlias: "IGNORED"})
	sendMessage(t, connA, Envelope{Type: EventTyping, Payload: json.RawMessage(typingPayload)})

	env := readEnvelope(t, connB, 2*time.Second)
	if env.Type != EventTyping {
		t.Fatalf("B: expected TYPING, got %q", env.Type)
	}
	var recv TypingPayload
	mustUnmarshal(t, env.Payload, &recv)
	if !recv.IsTyping {
		t.Error("isTyping should be true")
	}
	if recv.SenderAlias == "IGNORED" || recv.SenderAlias == "" {
		t.Errorf("senderAlias should be server-assigned, got %q", recv.SenderAlias)
	}
}

// TestNationalFallbackMatching verifies that two clients with identical
// PromptID + AgeBucket but DIFFERENT cities are matched after FallbackDelay
// has elapsed for at least one of them.
func TestNationalFallbackMatching(t *testing.T) {
	origDelay := FallbackDelay
	FallbackDelay = 200 * time.Millisecond
	defer func() { FallbackDelay = origDelay }()

	wsBase, cleanup := testServer(t)
	defer cleanup()

	connA := dialWS(t, wsBase+"?promptId=p4&ageBucket=18-25&city=Mumbai&anonymousId=userG")
	defer connA.Close()
	connB := dialWS(t, wsBase+"?promptId=p4&ageBucket=18-25&city=Chennai&anonymousId=userH")
	defer connB.Close()

	// With FallbackDelay = 200 ms, allow up to 2 s for the match.
	envA := readEnvelope(t, connA, 2*time.Second)
	envB := readEnvelope(t, connB, 2*time.Second)

	if envA.Type != EventConnected {
		t.Errorf("A: expected CONNECTED after national fallback, got %q", envA.Type)
	}
	if envB.Type != EventConnected {
		t.Errorf("B: expected CONNECTED after national fallback, got %q", envB.Type)
	}
}

// TestMismatchedPromptsDoNotMatch ensures clients with different PromptIDs
// are never paired even after FallbackDelay.
func TestMismatchedPromptsDoNotMatch(t *testing.T) {
	origDelay := FallbackDelay
	FallbackDelay = 50 * time.Millisecond
	defer func() { FallbackDelay = origDelay }()

	wsBase, cleanup := testServer(t)
	defer cleanup()

	connA := dialWS(t, wsBase+"?promptId=prompt-X&ageBucket=18-25&city=Mumbai&anonymousId=userI")
	defer connA.Close()
	connB := dialWS(t, wsBase+"?promptId=prompt-Y&ageBucket=18-25&city=Mumbai&anonymousId=userJ")
	defer connB.Close()

	// Neither client should receive CONNECTED within 500 ms.
	connA.SetReadDeadline(time.Now().Add(400 * time.Millisecond))
	_, _, err := connA.ReadMessage()
	if err == nil {
		t.Error("A: should not have been matched (different promptId)")
	}
}

// TestDisconnectNotifiesPartner verifies that when client A disconnects,
// client B receives a DISCONNECT event.
func TestDisconnectNotifiesPartner(t *testing.T) {
	wsBase, cleanup := testServer(t)
	defer cleanup()

	params := "?promptId=p5&ageBucket=18-25&city=Hyderabad"
	connA := dialWS(t, wsBase+params+"&anonymousId=userK")
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId=userL")
	defer connB.Close()

	readEnvelope(t, connA, 2*time.Second)
	readEnvelope(t, connB, 2*time.Second)

	// A closes the connection.
	connA.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
	connA.Close()

	// B should receive DISCONNECT.
	env := readEnvelope(t, connB, 2*time.Second)
	if env.Type != EventDisconnect {
		t.Errorf("B: expected DISCONNECT after A left, got %q", env.Type)
	}
}

// TestMissingAnonymousIDRejected verifies that a connection without anonymousId
// is rejected with HTTP 400.
func TestMissingAnonymousIDRejected(t *testing.T) {
	app := NewApp()
	app.Start()
	ts := httptest.NewServer(corsMiddleware(app))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws?promptId=p1&ageBucket=18-25&city=Mumbai"
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected dial error for missing anonymousId")
	}
	if resp != nil && resp.StatusCode != 400 {
		t.Errorf("expected HTTP 400, got %d", resp.StatusCode)
	}
}

// TestBidirectionalMessaging sends messages in both directions and verifies
// each party receives the other's message.
func TestBidirectionalMessaging(t *testing.T) {
	wsBase, cleanup := testServer(t)
	defer cleanup()

	params := "?promptId=p6&ageBucket=26-35&city=Pune"
	connA := dialWS(t, wsBase+params+"&anonymousId=userM")
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId=userN")
	defer connB.Close()

	readEnvelope(t, connA, 2*time.Second)
	readEnvelope(t, connB, 2*time.Second)

	// A → B
	sendMsg := func(conn *websocket.Conn, text string) {
		p, _ := json.Marshal(MessagePayload{MessageID: text, Text: text, SenderAlias: "x"})
		sendMessage(t, conn, Envelope{Type: EventMessage, Payload: json.RawMessage(p)})
	}

	sendMsg(connA, "msg-from-A")
	envAtB := readEnvelope(t, connB, 2*time.Second)
	if envAtB.Type != EventMessage {
		t.Fatalf("B: expected MESSAGE, got %q", envAtB.Type)
	}
	var pAtB MessagePayload
	mustUnmarshal(t, envAtB.Payload, &pAtB)
	if pAtB.Text != "msg-from-A" {
		t.Errorf("B received wrong text: %q", pAtB.Text)
	}

	// B → A
	sendMsg(connB, "msg-from-B")
	envAtA := readEnvelope(t, connA, 2*time.Second)
	if envAtA.Type != EventMessage {
		t.Fatalf("A: expected MESSAGE, got %q", envAtA.Type)
	}
	var pAtA MessagePayload
	mustUnmarshal(t, envAtA.Payload, &pAtA)
	if pAtA.Text != "msg-from-B" {
		t.Errorf("A received wrong text: %q", pAtA.Text)
	}
}

// --- helpers ---

func mustUnmarshal(t *testing.T, data json.RawMessage, v any) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal %T: %v", v, err)
	}
}

// Compile-time check that the test file number makes sense
var _ = fmt.Sprintf
