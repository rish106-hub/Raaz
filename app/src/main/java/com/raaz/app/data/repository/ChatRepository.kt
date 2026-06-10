package com.raaz.app.data.repository

import com.raaz.app.data.models.ChatMessage
import com.raaz.app.data.websocket.WebSocketClient
import com.raaz.app.data.websocket.WebSocketEvent
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.filterIsInstance
import kotlinx.coroutines.flow.map
import java.util.UUID

data class TypingState(
    val senderAlias: String,
    val isTyping: Boolean
)

data class ExtensionRequest(
    val requesterId: String,
    val requesterAlias: String
)

data class ExtensionResponse(
    val approved: Boolean,
    val responderAlias: String
)

class ChatRepository(
    private val webSocketClient: WebSocketClient,
    private val currentUserAlias: String
) {
    fun getMessages(): Flow<ChatMessage> {
        return webSocketClient.events
            .filterIsInstance<WebSocketEvent.Message>()
            .map { event ->
                ChatMessage(
                    messageId = event.messageId,
                    text = event.text,
                    senderAlias = event.senderAlias,
                    isOwnMessage = event.senderAlias == currentUserAlias,
                    timestamp = event.timestamp
                )
            }
    }

    fun getTypingState(): Flow<TypingState> {
        return webSocketClient.events
            .filterIsInstance<WebSocketEvent.Typing>()
            .map { event ->
                TypingState(
                    senderAlias = event.senderAlias,
                    isTyping = event.isTyping
                )
            }
    }

    fun sendMessage(text: String) {
        val messageId = UUID.randomUUID().toString()
        val event = WebSocketEvent.Message(
            messageId = messageId,
            text = text,
            senderAlias = currentUserAlias,
            timestamp = System.currentTimeMillis()
        )
        webSocketClient.send(event)
    }

    fun sendTyping(isTyping: Boolean) {
        val event = WebSocketEvent.Typing(
            isTyping = isTyping,
            senderAlias = currentUserAlias
        )
        webSocketClient.send(event)
    }

    fun getExtensionRequests(): Flow<ExtensionRequest> {
        return webSocketClient.events
            .filterIsInstance<WebSocketEvent.ExtensionRequest>()
            .map { event ->
                ExtensionRequest(
                    requesterId = event.requesterId,
                    requesterAlias = event.requesterAlias
                )
            }
    }

    fun getExtensionResponses(): Flow<ExtensionResponse> {
        return webSocketClient.events
            .filterIsInstance<WebSocketEvent.ExtensionResponse>()
            .map { event ->
                ExtensionResponse(
                    approved = event.approved,
                    responderAlias = event.responderAlias
                )
            }
    }

    fun sendExtensionRequest(requesterId: String) {
        val event = WebSocketEvent.ExtensionRequest(
            requesterId = requesterId,
            requesterAlias = currentUserAlias
        )
        webSocketClient.send(event)
    }

    fun sendExtensionResponse(approved: Boolean) {
        val event = WebSocketEvent.ExtensionResponse(
            approved = approved,
            responderAlias = currentUserAlias
        )
        webSocketClient.send(event)
    }

    fun disconnect() {
        webSocketClient.disconnect()
    }
}
