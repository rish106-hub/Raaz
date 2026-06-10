package main

import (
	"encoding/json"
	"log"
	"time"
)

// routeMessage is called by the relay before forwarding a frame to the partner.
// It validates the envelope, delegates to the appropriate handler, and returns
// the final bytes to transmit. Returns nil to drop the frame silently.
func routeMessage(raw []byte, src *Client) []byte {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		log.Printf("invalid JSON from %s: %v", src.alias, err)
		return nil
	}

	switch env.Type {
	case EventMessage:
		return handleChatMessage(env.Payload, src)
	case EventTyping:
		return handleTyping(env.Payload, src)
	case EventExtensionRequest:
		return handleExtensionRequest(env.Payload, src)
	case EventExtensionResponse:
		return handleExtensionResponse(env.Payload, src)
	case EventHandleExchange:
		return handleHandleExchange(env.Payload, src)
	case EventHandleRevealed:
		// Forwarded verbatim: payload is server-opaque (handle string not
		// inspected here; the Android client controls reveal logic).
		return raw
	default:
		// Unknown types dropped; avoids forwarding unexpected client frames.
		log.Printf("unknown event type %q from %s", env.Type, src.alias)
		return nil
	}
}

func handleChatMessage(payload json.RawMessage, src *Client) []byte {
	var p MessagePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil
	}
	// Server enforces sender identity and timestamps.
	p.SenderAlias = src.alias
	p.Timestamp = time.Now().UnixMilli()
	out, err := marshalEnvelope(EventMessage, p)
	if err != nil {
		return nil
	}
	return out
}

func handleTyping(payload json.RawMessage, src *Client) []byte {
	var p TypingPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil
	}
	p.SenderAlias = src.alias
	out, err := marshalEnvelope(EventTyping, p)
	if err != nil {
		return nil
	}
	return out
}

func handleExtensionRequest(payload json.RawMessage, src *Client) []byte {
	var p ExtensionRequestPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil
	}
	p.RequesterAlias = src.alias
	out, err := marshalEnvelope(EventExtensionRequest, p)
	if err != nil {
		return nil
	}
	return out
}

func handleExtensionResponse(payload json.RawMessage, src *Client) []byte {
	var p ExtensionResponsePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil
	}
	p.ResponderAlias = src.alias
	out, err := marshalEnvelope(EventExtensionResponse, p)
	if err != nil {
		return nil
	}
	return out
}

func handleHandleExchange(payload json.RawMessage, src *Client) []byte {
	var p HandleExchangePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil
	}
	p.UserAlias = src.alias
	out, err := marshalEnvelope(EventHandleExchange, p)
	if err != nil {
		return nil
	}
	return out
}
