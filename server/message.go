package main

import "encoding/json"

// EventType is the wire type discriminator for all WebSocket messages.
type EventType string

const (
	EventConnected         EventType = "CONNECTED"
	EventMessage           EventType = "MESSAGE"
	EventTyping            EventType = "TYPING"
	EventExtensionRequest  EventType = "EXTENSION_REQUEST"
	EventExtensionResponse EventType = "EXTENSION_RESPONSE"
	EventHandleExchange    EventType = "HANDLE_EXCHANGE"
	EventHandleRevealed    EventType = "HANDLE_REVEALED"
	EventError             EventType = "ERROR"
	EventDisconnect        EventType = "DISCONNECT"
	EventModerationAlert   EventType = "MODERATION_ALERT"
	EventCrisisTriggered   EventType = "CRISIS_TRIGGERED"
)

// Envelope is the top-level frame for every WebSocket message in both directions.
type Envelope struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// marshalEnvelope serialises a typed envelope ready for wire transmission.
func marshalEnvelope(t EventType, payload any) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: t, Payload: json.RawMessage(p)})
}

// --- Payload structs (server → client) ---

type ConnectedPayload struct {
	MatchID                string `json:"matchId"`
	PartnerAlias           string `json:"partnerAlias"`
	SessionDurationSeconds int64  `json:"sessionDurationSeconds"`
}

type ErrorPayload struct {
	Message string  `json:"message"`
	Code    *string `json:"code,omitempty"`
}

type DisconnectPayload struct {
	Reason string `json:"reason"`
}

// --- Payload structs (client → server, also routed to partner) ---

type MessagePayload struct {
	MessageID   string `json:"messageId"`
	Text        string `json:"text"`
	SenderAlias string `json:"senderAlias"`
	Timestamp   int64  `json:"timestamp"`
}

type TypingPayload struct {
	IsTyping    bool   `json:"isTyping"`
	SenderAlias string `json:"senderAlias"`
}

type ExtensionRequestPayload struct {
	RequesterID    string `json:"requesterId"`
	RequesterAlias string `json:"requesterAlias"`
}

type ExtensionResponsePayload struct {
	Approved       bool   `json:"approved"`
	ResponderAlias string `json:"responderAlias"`
}

type HandleExchangePayload struct {
	UserID    string `json:"userId"`
	UserAlias string `json:"userAlias"`
	Approved  bool   `json:"approved"`
}

type HandleRevealedPayload struct {
	PartnerHandle string `json:"partnerHandle"`
}

// ModerationAlertPayload is sent to the offending client when a message is blocked.
type ModerationAlertPayload struct {
	Category  string `json:"category"`
	StrikeNum int    `json:"strikeNum"`
	Action    string `json:"action"` // "warning" or "disconnect"
	Reason    string `json:"reason"`
}

// CrisisTriggeredPayload is broadcast to both session participants on crisis detection.
type CrisisTriggeredPayload struct {
	Helplines []string `json:"helplines"`
	Message   string   `json:"message"`
}

// RegistrationParams are extracted from WebSocket connection URL query parameters.
type RegistrationParams struct {
	PromptID    string
	AgeBucket   string
	City        string
	AnonymousID string
}
