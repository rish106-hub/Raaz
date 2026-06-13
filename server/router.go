package main

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/raaz/server/store"
)

// routeMessage is called by the relay before forwarding a frame to the partner.
// It validates the envelope, runs moderation on chat messages, and delegates to
// the appropriate handler. Returns nil to drop the frame silently.
func routeMessage(raw []byte, src *Client, sess *Session) []byte {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		slog.Warn("invalid JSON from client", "alias", src.alias, "err", err)
		return nil
	}

	switch env.Type {
	case EventMessage:
		return handleChatMessage(env.Payload, src, sess)
	case EventTyping:
		return handleTyping(env.Payload, src)
	case EventExtensionRequest:
		return handleExtensionRequest(env.Payload, src)
	case EventExtensionResponse:
		return handleExtensionResponse(env.Payload, src, sess)
	case EventHandleExchange:
		return handleHandleExchange(env.Payload, src)
	case EventHandleRevealed:
		return raw
	default:
		slog.Warn("unknown event type", "type", env.Type, "alias", src.alias)
		return nil
	}
}

func handleChatMessage(payload json.RawMessage, src *Client, sess *Session) []byte {
	var p MessagePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil
	}
	p.SenderAlias = src.alias
	p.Timestamp = time.Now().UnixMilli()

	result := Moderate(p.Text)

	if result.IsCrisis {
		crisisTriggersTotal.Inc()
		sess.triggerCrisis()
		return nil
	}

	if result.Flagged {
		moderationStrikesTotal.WithLabelValues(string(result.Category)).Inc()
		strike, err := globalStrikes.RecordStrike(src.params.AnonymousID)
		if err != nil {
			slog.Error("record strike error", "anonymousID", src.params.AnonymousID, "err", err)
			return nil
		}
		notifyModerationAlert(src, result, strike)
		if strike.Action == store.StrikeActionDisconnect {
			go func() {
				time.Sleep(500 * time.Millisecond)
				src.conn.Close()
			}()
		}
		return nil
	}

	out, err := marshalEnvelope(EventMessage, p)
	if err != nil {
		return nil
	}
	return out
}

func notifyModerationAlert(c *Client, mod ModerationResult, strike store.StrikeResult) {
	action := "warning"
	if strike.Action == store.StrikeActionDisconnect {
		action = "disconnect"
	}
	data, err := marshalEnvelope(EventModerationAlert, ModerationAlertPayload{
		Category:  string(mod.Category),
		StrikeNum: strike.Strikes,
		Action:    action,
		Reason:    "Your message violated our community guidelines.",
	})
	if err != nil {
		return
	}
	c.safeSend(data)
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

// handleExtensionResponse enforces max-1 extension per session (M-2).
// A second approved extension response is silently dropped.
func handleExtensionResponse(payload json.RawMessage, src *Client, sess *Session) []byte {
	var p ExtensionResponsePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil
	}
	p.ResponderAlias = src.alias
	if p.Approved {
		if sess.extensions.Add(1) > 1 {
			return nil // second extension blocked server-side
		}
	}
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
