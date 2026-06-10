package com.raaz.app.data.websocket

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
sealed class WebSocketEvent {
    @Serializable
    @SerialName("CONNECTED")
    data class Connected(
        val matchId: String,
        val partnerAlias: String,
        val sessionDurationSeconds: Long
    ) : WebSocketEvent()

    @Serializable
    @SerialName("MESSAGE")
    data class Message(
        val messageId: String,
        val text: String,
        val senderAlias: String,
        val timestamp: Long
    ) : WebSocketEvent()

    @Serializable
    @SerialName("TYPING")
    data class Typing(
        val isTyping: Boolean,
        val senderAlias: String
    ) : WebSocketEvent()

    @Serializable
    @SerialName("EXTENSION_REQUEST")
    data class ExtensionRequest(
        val requesterId: String,
        val requesterAlias: String
    ) : WebSocketEvent()

    @Serializable
    @SerialName("EXTENSION_RESPONSE")
    data class ExtensionResponse(
        val approved: Boolean,
        val responderAlias: String
    ) : WebSocketEvent()

    @Serializable
    @SerialName("HANDLE_EXCHANGE")
    data class HandleExchange(
        val userId: String,
        val userAlias: String,
        val approved: Boolean
    ) : WebSocketEvent()

    @Serializable
    @SerialName("HANDLE_REVEALED")
    data class HandleRevealed(
        val partnerHandle: String
    ) : WebSocketEvent()

    @Serializable
    @SerialName("ERROR")
    data class Error(
        val message: String,
        val code: String? = null
    ) : WebSocketEvent()

    @Serializable
    @SerialName("DISCONNECT")
    data class Disconnect(
        val reason: String
    ) : WebSocketEvent()
}

data class WebSocketMessage(
    val type: String,
    val payload: String
)
