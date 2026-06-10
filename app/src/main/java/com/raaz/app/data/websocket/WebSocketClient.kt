package com.raaz.app.data.websocket

import android.util.Log
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import java.util.concurrent.TimeUnit

class WebSocketClient(
    private val url: String,
    private val okHttpClient: OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(10, TimeUnit.SECONDS)
        .readTimeout(10, TimeUnit.SECONDS)
        .writeTimeout(10, TimeUnit.SECONDS)
        .build()
) {
    private var webSocket: WebSocket? = null
    private val _events = MutableSharedFlow<WebSocketEvent>(replay = 1)
    val events: Flow<WebSocketEvent> = _events.asSharedFlow()

    private var isConnected = false
    private var reconnectAttempts = 0
    private val maxReconnectAttempts = 5

    fun connect() {
        if (isConnected || reconnectAttempts > maxReconnectAttempts) return

        val request = Request.Builder()
            .url(url)
            .build()

        webSocket = okHttpClient.newWebSocket(request, object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: okhttp3.Response) {
                Log.d("WebSocket", "Connected")
                isConnected = true
                reconnectAttempts = 0
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                try {
                    val event = parseWebSocketMessage(text)
                    _events.tryEmit(event)
                } catch (e: Exception) {
                    Log.e("WebSocket", "Error parsing message: ${e.message}")
                }
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: okhttp3.Response?) {
                Log.e("WebSocket", "Connection failed: ${t.message}")
                isConnected = false
                attemptReconnect()
            }

            override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                Log.d("WebSocket", "Closing: $reason")
                isConnected = false
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                Log.d("WebSocket", "Closed: $reason")
                isConnected = false
            }
        })
    }

    fun send(event: WebSocketEvent) {
        if (!isConnected) {
            Log.w("WebSocket", "Cannot send: not connected")
            return
        }

        try {
            val json = serializeEvent(event)
            webSocket?.send(json)
        } catch (e: Exception) {
            Log.e("WebSocket", "Error sending event: ${e.message}")
        }
    }

    fun disconnect() {
        isConnected = false
        webSocket?.close(1000, "User disconnect")
        webSocket = null
    }

    private fun attemptReconnect() {
        if (reconnectAttempts < maxReconnectAttempts) {
            reconnectAttempts++
            val delay = (1000 * Math.pow(2.0, reconnectAttempts.toDouble())).toLong()
            Log.d("WebSocket", "Reconnect attempt $reconnectAttempts after ${delay}ms")
            Thread {
                Thread.sleep(delay)
                connect()
            }.start()
        }
    }

    private fun parseWebSocketMessage(json: String): WebSocketEvent {
        return try {
            when {
                json.contains("\"type\":\"CONNECTED\"") -> {
                    val matchId = extractString(json, "matchId")
                    val partnerAlias = extractString(json, "partnerAlias")
                    val sessionDurationSeconds = extractLong(json, "sessionDurationSeconds")
                    WebSocketEvent.Connected(matchId, partnerAlias, sessionDurationSeconds)
                }
                json.contains("\"type\":\"MESSAGE\"") -> {
                    val messageId = extractString(json, "messageId")
                    val text = extractString(json, "text")
                    val senderAlias = extractString(json, "senderAlias")
                    val timestamp = extractLong(json, "timestamp")
                    WebSocketEvent.Message(messageId, text, senderAlias, timestamp)
                }
                json.contains("\"type\":\"TYPING\"") -> {
                    val isTyping = extractBoolean(json, "isTyping")
                    val senderAlias = extractString(json, "senderAlias")
                    WebSocketEvent.Typing(isTyping, senderAlias)
                }
                json.contains("\"type\":\"EXTENSION_REQUEST\"") -> {
                    val requesterId = extractString(json, "requesterId")
                    val requesterAlias = extractString(json, "requesterAlias")
                    WebSocketEvent.ExtensionRequest(requesterId, requesterAlias)
                }
                json.contains("\"type\":\"EXTENSION_RESPONSE\"") -> {
                    val approved = extractBoolean(json, "approved")
                    val responderAlias = extractString(json, "responderAlias")
                    WebSocketEvent.ExtensionResponse(approved, responderAlias)
                }
                json.contains("\"type\":\"HANDLE_EXCHANGE\"") -> {
                    val userId = extractString(json, "userId")
                    val userAlias = extractString(json, "userAlias")
                    val approved = extractBoolean(json, "approved")
                    WebSocketEvent.HandleExchange(userId, userAlias, approved)
                }
                json.contains("\"type\":\"HANDLE_REVEALED\"") -> {
                    val partnerHandle = extractString(json, "partnerHandle")
                    WebSocketEvent.HandleRevealed(partnerHandle)
                }
                json.contains("\"type\":\"ERROR\"") -> {
                    val message = extractString(json, "message")
                    val code = extractStringOrNull(json, "code")
                    WebSocketEvent.Error(message, code)
                }
                json.contains("\"type\":\"DISCONNECT\"") -> {
                    val reason = extractString(json, "reason")
                    WebSocketEvent.Disconnect(reason)
                }
                else -> throw IllegalArgumentException("Unknown event type")
            }
        } catch (e: Exception) {
            Log.e("WebSocket", "Parse error: ${e.message}")
            WebSocketEvent.Error("Parse error: ${e.message}")
        }
    }

    private fun serializeEvent(event: WebSocketEvent): String {
        return when (event) {
            is WebSocketEvent.Message -> {
                """{"type":"MESSAGE","payload":{"messageId":"${event.messageId}","text":"${escapeJson(event.text)}","senderAlias":"${event.senderAlias}","timestamp":${event.timestamp}}}"""
            }
            is WebSocketEvent.Typing -> {
                """{"type":"TYPING","payload":{"isTyping":${event.isTyping},"senderAlias":"${event.senderAlias}"}}"""
            }
            is WebSocketEvent.ExtensionRequest -> {
                """{"type":"EXTENSION_REQUEST","payload":{"requesterId":"${event.requesterId}","requesterAlias":"${event.requesterAlias}"}}"""
            }
            is WebSocketEvent.ExtensionResponse -> {
                """{"type":"EXTENSION_RESPONSE","payload":{"approved":${event.approved},"responderAlias":"${event.responderAlias}"}}"""
            }
            is WebSocketEvent.HandleExchange -> {
                """{"type":"HANDLE_EXCHANGE","payload":{"userId":"${event.userId}","userAlias":"${event.userAlias}","approved":${event.approved}}}"""
            }
            else -> "{}"
        }
    }

    private fun extractString(json: String, key: String): String {
        val regex = """"$key"\s*:\s*"([^"]*)"""".toRegex()
        return regex.find(json)?.groupValues?.get(1) ?: ""
    }

    private fun extractStringOrNull(json: String, key: String): String? {
        val regex = """"$key"\s*:\s*"([^"]*)"""".toRegex()
        return regex.find(json)?.groupValues?.get(1)
    }

    private fun extractLong(json: String, key: String): Long {
        val regex = """"$key"\s*:\s*(\d+)""".toRegex()
        return regex.find(json)?.groupValues?.get(1)?.toLongOrNull() ?: 0L
    }

    private fun extractBoolean(json: String, key: String): Boolean {
        val regex = """"$key"\s*:\s*(true|false)""".toRegex()
        return regex.find(json)?.groupValues?.get(1) == "true"
    }

    private fun escapeJson(text: String): String {
        return text
            .replace("\\", "\\\\")
            .replace("\"", "\\\"")
            .replace("\n", "\\n")
            .replace("\r", "\\r")
    }
}
