package com.raaz.app.data.websocket

// Plain sealed class hierarchy — parsed manually by WebSocketClient via Gson.
// No kotlinx.serialization dependency needed or used.
sealed class WebSocketEvent {

    data class Registered(
        val connToken: String
    ) : WebSocketEvent()

    data class Connected(
        val matchId: String,
        val partnerAlias: String,
        val sessionDurationSeconds: Long
    ) : WebSocketEvent()

    data class Message(
        val messageId: String,
        val text: String,
        val senderAlias: String,
        val timestamp: Long
    ) : WebSocketEvent()

    data class Typing(
        val isTyping: Boolean,
        val senderAlias: String
    ) : WebSocketEvent()

    data class ExtensionRequest(
        val requesterId: String,
        val requesterAlias: String
    ) : WebSocketEvent()

    data class ExtensionResponse(
        val approved: Boolean,
        val responderAlias: String
    ) : WebSocketEvent()

    data class HandleExchange(
        val userId: String,
        val userAlias: String,
        val approved: Boolean
    ) : WebSocketEvent()

    data class HandleRevealed(
        val partnerHandle: String
    ) : WebSocketEvent()

    data class Error(
        val message: String,
        val code: String? = null
    ) : WebSocketEvent()

    data class Disconnect(
        val reason: String
    ) : WebSocketEvent()

    data class ModerationAlert(
        val category: String,
        val strikeNum: Int,
        val action: String,
        val reason: String
    ) : WebSocketEvent()

    data class CrisisTriggered(
        val helplines: List<String>,
        val message: String
    ) : WebSocketEvent()
}
