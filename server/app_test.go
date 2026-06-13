package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// drainRegistered reads and discards the REGISTERED envelope that the server sends
// immediately after every successful WebSocket upgrade. Must be called once per
// connection before reading any subsequent events (CONNECTED, MODERATION_ALERT, etc.).
func drainRegistered(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	env := readEnvelope(t, conn, 2*time.Second)
	if env.Type != EventRegistered {
		t.Fatalf("expected REGISTERED on connect, got %q", env.Type)
	}
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

	params := "?promptId=p1&ageBucket=18-22&city=Mumbai"

	// Connect both clients before reading from either so the match loop
	// can pair them after both are in the queue.
	connA := dialWS(t, wsBase+params+"&anonymousId=userA")
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId=userB")
	defer connB.Close()

	drainRegistered(t, connA)
	drainRegistered(t, connB)

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

	params := "?promptId=p2&ageBucket=23-28&city=Delhi"
	connA := dialWS(t, wsBase+params+"&anonymousId=userC")
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId=userD")
	defer connB.Close()

	drainRegistered(t, connA)
	drainRegistered(t, connB)

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

	params := "?promptId=p3&ageBucket=18-22&city=Bangalore"
	connA := dialWS(t, wsBase+params+"&anonymousId=userE")
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId=userF")
	defer connB.Close()

	drainRegistered(t, connA)
	drainRegistered(t, connB)

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

	connA := dialWS(t, wsBase+"?promptId=p4&ageBucket=18-22&city=Mumbai&anonymousId=userG")
	defer connA.Close()
	connB := dialWS(t, wsBase+"?promptId=p4&ageBucket=18-22&city=Chennai&anonymousId=userH")
	defer connB.Close()

	drainRegistered(t, connA)
	drainRegistered(t, connB)

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

	connA := dialWS(t, wsBase+"?promptId=prompt-X&ageBucket=18-22&city=Mumbai&anonymousId=userI")
	defer connA.Close()
	connB := dialWS(t, wsBase+"?promptId=prompt-Y&ageBucket=18-22&city=Mumbai&anonymousId=userJ")
	defer connB.Close()

	// Drain the REGISTERED event first; then verify no CONNECTED arrives.
	drainRegistered(t, connA)
	drainRegistered(t, connB)

	// Neither client should receive CONNECTED within 400 ms.
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

	params := "?promptId=p5&ageBucket=18-22&city=Hyderabad"
	connA := dialWS(t, wsBase+params+"&anonymousId=userK")
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId=userL")
	defer connB.Close()

	drainRegistered(t, connA)
	drainRegistered(t, connB)
	readEnvelope(t, connA, 2*time.Second) // CONNECTED
	readEnvelope(t, connB, 2*time.Second) // CONNECTED

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

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws?promptId=p1&ageBucket=18-22&city=Mumbai"
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

	params := "?promptId=p6&ageBucket=23-28&city=Pune"
	connA := dialWS(t, wsBase+params+"&anonymousId=userM")
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId=userN")
	defer connB.Close()

	drainRegistered(t, connA)
	drainRegistered(t, connB)
	readEnvelope(t, connA, 2*time.Second) // CONNECTED
	readEnvelope(t, connB, 2*time.Second) // CONNECTED

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

// TestModerationBlocksBannedPhrase verifies that a harassment phrase sent by
// client A is blocked (B never receives it) and A receives a MODERATION_ALERT
// with action=warning and category=harassment.
func TestModerationBlocksBannedPhrase(t *testing.T) {
	wsBase, cleanup := testServer(t)
	defer cleanup()

	aID := t.Name() + "-A"
	bID := t.Name() + "-B"
	params := "?promptId=mod-p1&ageBucket=18-22&city=Delhi"

	connA := dialWS(t, wsBase+params+"&anonymousId="+aID)
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId="+bID)
	defer connB.Close()

	drainRegistered(t, connA)
	drainRegistered(t, connB)
	readEnvelope(t, connA, 2*time.Second) // CONNECTED
	readEnvelope(t, connB, 2*time.Second) // CONNECTED

	sendViolation(t, connA, "kill yourself")

	// A receives MODERATION_ALERT — strike 1, action=warning.
	env := readEnvelope(t, connA, 2*time.Second)
	if env.Type != EventModerationAlert {
		t.Fatalf("A: expected MODERATION_ALERT, got %q", env.Type)
	}
	var alert ModerationAlertPayload
	mustUnmarshal(t, env.Payload, &alert)
	if alert.Action != "warning" {
		t.Errorf("strike 1 action: want warning, got %q", alert.Action)
	}
	if alert.StrikeNum != 1 {
		t.Errorf("strikeNum: want 1, got %d", alert.StrikeNum)
	}
	if alert.Category != string(CategoryHarassment) {
		t.Errorf("category: want %q, got %q", CategoryHarassment, alert.Category)
	}

	// B must NOT receive the blocked message — a short read deadline should expire.
	connB.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	_, _, err := connB.ReadMessage()
	if err == nil {
		t.Error("B should not receive a moderation-blocked message")
	}
}

// TestStrikeSystemDisconnectsAndBans verifies that two violations from the same
// anonymousId trigger a disconnect on the second strike, notify the partner with
// DISCONNECT, and block the user from reconnecting (HTTP 403).
func TestStrikeSystemDisconnectsAndBans(t *testing.T) {
	wsBase, cleanup := testServer(t)
	defer cleanup()

	aID := t.Name() + "-A"
	bID := t.Name() + "-B"
	params := "?promptId=mod-p2&ageBucket=18-22&city=Mumbai"

	connA := dialWS(t, wsBase+params+"&anonymousId="+aID)
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId="+bID)
	defer connB.Close()

	drainRegistered(t, connA)
	drainRegistered(t, connB)
	readEnvelope(t, connA, 2*time.Second) // CONNECTED
	readEnvelope(t, connB, 2*time.Second) // CONNECTED

	// Strike 1: warning.
	sendViolation(t, connA, "kill yourself")
	env := readEnvelope(t, connA, 2*time.Second)
	if env.Type != EventModerationAlert {
		t.Fatalf("A: expected MODERATION_ALERT on strike 1, got %q", env.Type)
	}

	// Strike 2: disconnect.
	sendViolation(t, connA, "you should die")
	env = readEnvelope(t, connA, 2*time.Second)
	if env.Type != EventModerationAlert {
		t.Fatalf("A: expected MODERATION_ALERT on strike 2, got %q", env.Type)
	}
	var alert ModerationAlertPayload
	mustUnmarshal(t, env.Payload, &alert)
	if alert.Action != "disconnect" {
		t.Errorf("strike 2 action: want disconnect, got %q", alert.Action)
	}
	if alert.StrikeNum != 2 {
		t.Errorf("strikeNum: want 2, got %d", alert.StrikeNum)
	}

	// B must receive DISCONNECT after the server force-closes A's connection.
	// Allow 3 s for the 500 ms server delay + 150 ms relay flush.
	env = readEnvelope(t, connB, 3*time.Second)
	if env.Type != EventDisconnect {
		t.Errorf("B: expected DISCONNECT after A banned, got %q", env.Type)
	}

	// Banned user must be rejected with HTTP 403.
	_, resp, err := websocket.DefaultDialer.Dial(wsBase+params+"&anonymousId="+aID, nil)
	if err == nil {
		t.Fatal("banned user should not be able to reconnect")
	}
	if resp != nil && resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected HTTP 403 for banned user, got %d", resp.StatusCode)
	}
}

// TestCrisisTriggerBroadcastsToBothClients verifies that a self-harm phrase
// causes CRISIS_TRIGGERED to be sent to both participants in the session.
func TestCrisisTriggerBroadcastsToBothClients(t *testing.T) {
	wsBase, cleanup := testServer(t)
	defer cleanup()

	aID := t.Name() + "-A"
	bID := t.Name() + "-B"
	params := "?promptId=mod-p3&ageBucket=18-22&city=Bangalore"

	connA := dialWS(t, wsBase+params+"&anonymousId="+aID)
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId="+bID)
	defer connB.Close()

	drainRegistered(t, connA)
	drainRegistered(t, connB)
	readEnvelope(t, connA, 2*time.Second) // CONNECTED
	readEnvelope(t, connB, 2*time.Second) // CONNECTED

	sendViolation(t, connA, "i want to kill myself")

	envA := readEnvelope(t, connA, 2*time.Second)
	envB := readEnvelope(t, connB, 2*time.Second)

	if envA.Type != EventCrisisTriggered {
		t.Errorf("A: expected CRISIS_TRIGGERED, got %q", envA.Type)
	}
	if envB.Type != EventCrisisTriggered {
		t.Errorf("B: expected CRISIS_TRIGGERED, got %q", envB.Type)
	}

	var crisis CrisisTriggeredPayload
	mustUnmarshal(t, envA.Payload, &crisis)
	if len(crisis.Helplines) == 0 {
		t.Error("CRISIS_TRIGGERED must include helplines")
	}
	if crisis.Message == "" {
		t.Error("CRISIS_TRIGGERED must include a message")
	}
}

// TestMessagesPausedDuringCrisis verifies that the session drops messages while
// paused and resumes delivery after CrisisPauseDuration elapses.
func TestMessagesPausedDuringCrisis(t *testing.T) {
	orig := CrisisPauseDuration
	CrisisPauseDuration = 400 * time.Millisecond
	defer func() { CrisisPauseDuration = orig }()

	wsBase, cleanup := testServer(t)
	defer cleanup()

	aID := t.Name() + "-A"
	bID := t.Name() + "-B"
	params := "?promptId=mod-p4&ageBucket=23-28&city=Chennai"

	connA := dialWS(t, wsBase+params+"&anonymousId="+aID)
	defer connA.Close()
	connB := dialWS(t, wsBase+params+"&anonymousId="+bID)
	defer connB.Close()

	drainRegistered(t, connA)
	drainRegistered(t, connB)
	readEnvelope(t, connA, 2*time.Second) // CONNECTED
	readEnvelope(t, connB, 2*time.Second) // CONNECTED

	// Trigger crisis; drain CRISIS_TRIGGERED from both sides.
	sendViolation(t, connA, "thinking about suicide")
	readEnvelope(t, connA, 2*time.Second)
	readEnvelope(t, connB, 2*time.Second)

	// While paused, send a message from A — the relay must drop it.
	// Avoid calling SetReadDeadline on connB: gorilla/websocket stores the first
	// read error and returns it on all subsequent reads, poisoning the connection.
	p, _ := json.Marshal(MessagePayload{MessageID: "paused", Text: "hello paused", SenderAlias: "x"})
	sendMessage(t, connA, Envelope{Type: EventMessage, Payload: json.RawMessage(p)})

	// Wait for the pause to expire, then send a second message.
	// Both messages are processed in order by the relay goroutine.
	// If the pause worked, only the second message reaches B.
	time.Sleep(CrisisPauseDuration + 150*time.Millisecond)

	p2, _ := json.Marshal(MessagePayload{MessageID: "resumed", Text: "back now", SenderAlias: "x"})
	sendMessage(t, connA, Envelope{Type: EventMessage, Payload: json.RawMessage(p2)})

	env := readEnvelope(t, connB, 2*time.Second)
	if env.Type != EventMessage {
		t.Errorf("B: expected MESSAGE after crisis resolved, got %q", env.Type)
	}
	var mp MessagePayload
	mustUnmarshal(t, env.Payload, &mp)
	// If pause did NOT work, B would receive "hello paused" first.
	if mp.Text != "back now" {
		t.Errorf("B received %q; pause may not have dropped the first message", mp.Text)
	}
}

// --- helpers ---

// sendViolation sends a chat MESSAGE envelope with the given text.
func sendViolation(t *testing.T, conn *websocket.Conn, text string) {
	t.Helper()
	p, _ := json.Marshal(MessagePayload{MessageID: "viol", Text: text, SenderAlias: "x"})
	sendMessage(t, conn, Envelope{Type: EventMessage, Payload: json.RawMessage(p)})
}

func mustUnmarshal(t *testing.T, data json.RawMessage, v any) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal %T: %v", v, err)
	}
}

// Compile-time check that the test file number makes sense
var _ = fmt.Sprintf
