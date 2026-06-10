package com.raaz.app.data.websocket

import android.util.Log
import com.google.gson.JsonParser
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import java.util.concurrent.TimeUnit
import kotlin.math.pow

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

    private val _connectionState = MutableStateFlow(ConnectionState.DISCONNECTED)
    val connectionState: StateFlow<ConnectionState> = _connectionState.asStateFlow()

    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
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
                _connectionState.value = ConnectionState.CONNECTED
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
                if (reconnectAttempts < maxReconnectAttempts) {
                    _connectionState.value = ConnectionState.RECONNECTING
                } else {
                    _connectionState.value = ConnectionState.DISCONNECTED
                }
                attemptReconnect()
            }

            override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                Log.d("WebSocket", "Closing: $reason")
                isConnected = false
                _connectionState.value = ConnectionState.DISCONNECTED
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                Log.d("WebSocket", "Closed: $reason")
                isConnected = false
                _connectionState.value = ConnectionState.DISCONNECTED
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
        _connectionState.value = ConnectionState.DISCONNECTED
        scope.cancel()
        webSocket?.close(1000, "User disconnect")
        webSocket = null
    }

    private fun attemptReconnect() {
        if (reconnectAttempts < maxReconnectAttempts) {
            reconnectAttempts++
            val delayMs = (1000 * 2.0.pow(reconnectAttempts.toDouble())).toLong()
            Log.d("WebSocket", "Reconnect attempt $reconnectAttempts after ${delayMs}ms")
            scope.launch {
                delay(delayMs)
                connect()
            }
        }
    }

    private fun parseWebSocketMessage(json: String): WebSocketEvent {
        return try {
            val root = JsonParser.parseString(json).asJsonObject
            val type = root.get("type")?.asString
                ?: throw IllegalArgumentException("Missing type field")
            val p = root.getAsJsonObject("payload") ?: root

            when (type) {
                "CONNECTED" -> WebSocketEvent.Connected(
                    matchId = p.get("matchId")?.asString ?: "",
                    partnerAlias = p.get("partnerAlias")?.asString ?: "",
                    sessionDurationSeconds = p.get("sessionDurationSeconds")?.asLong ?: 0L
                )
                "MESSAGE" -> WebSocketEvent.Message(
                    messageId = p.get("messageId")?.asString ?: "",
                    text = p.get("text")?.asString ?: "",
                    senderAlias = p.get("senderAlias")?.asString ?: "",
                    timestamp = p.get("timestamp")?.asLong ?: 0L
                )
                "TYPING" -> WebSocketEvent.Typing(
                    isTyping = p.get("isTyping")?.asBoolean ?: false,
                    senderAlias = p.get("senderAlias")?.asString ?: ""
                )
                "EXTENSION_REQUEST" -> WebSocketEvent.ExtensionRequest(
                    requesterId = p.get("requesterId")?.asString ?: "",
                    requesterAlias = p.get("requesterAlias")?.asString ?: ""
                )
                "EXTENSION_RESPONSE" -> WebSocketEvent.ExtensionResponse(
                    approved = p.get("approved")?.asBoolean ?: false,
                    responderAlias = p.get("responderAlias")?.asString ?: ""
                )
                "HANDLE_EXCHANGE" -> WebSocketEvent.HandleExchange(
                    userId = p.get("userId")?.asString ?: "",
                    userAlias = p.get("userAlias")?.asString ?: "",
                    approved = p.get("approved")?.asBoolean ?: false
                )
                "HANDLE_REVEALED" -> WebSocketEvent.HandleRevealed(
                    partnerHandle = p.get("partnerHandle")?.asString ?: ""
                )
                "ERROR" -> WebSocketEvent.Error(
                    message = p.get("message")?.asString ?: "Unknown error",
                    code = if (p.has("code") && !p.get("code").isJsonNull) p.get("code").asString else null
                )
                "DISCONNECT" -> WebSocketEvent.Disconnect(
                    reason = p.get("reason")?.asString ?: ""
                )
                else -> throw IllegalArgumentException("Unknown event type: $type")
            }
        } catch (e: Exception) {
            Log.e("WebSocket", "Parse error: ${e.message}")
            WebSocketEvent.Error("Parse error: ${e.message}")
        }
    }

    private fun serializeEvent(event: WebSocketEvent): String {
        return when (event) {
            is WebSocketEvent.Message ->
                """{"type":"MESSAGE","payload":{"messageId":"${event.messageId}","text":"${escapeJson(event.text)}","senderAlias":"${event.senderAlias}","timestamp":${event.timestamp}}}"""
            is WebSocketEvent.Typing ->
                """{"type":"TYPING","payload":{"isTyping":${event.isTyping},"senderAlias":"${event.senderAlias}"}}"""
            is WebSocketEvent.ExtensionRequest ->
                """{"type":"EXTENSION_REQUEST","payload":{"requesterId":"${event.requesterId}","requesterAlias":"${event.requesterAlias}"}}"""
            is WebSocketEvent.ExtensionResponse ->
                """{"type":"EXTENSION_RESPONSE","payload":{"approved":${event.approved},"responderAlias":"${event.responderAlias}"}}"""
            is WebSocketEvent.HandleExchange ->
                """{"type":"HANDLE_EXCHANGE","payload":{"userId":"${event.userId}","userAlias":"${event.userAlias}","approved":${event.approved}}}"""
            is WebSocketEvent.Connected ->
                """{"type":"CONNECTED","payload":{"matchId":"${event.matchId}","partnerAlias":"${event.partnerAlias}","sessionDurationSeconds":${event.sessionDurationSeconds}}}"""
            is WebSocketEvent.Error ->
                """{"type":"ERROR","payload":{"message":"${escapeJson(event.message)}","code":${event.code?.let { "\"$it\"" } ?: "null"}}}"""
            is WebSocketEvent.Disconnect ->
                """{"type":"DISCONNECT","payload":{"reason":"${escapeJson(event.reason)}"}}"""
            else -> "{}"
        }
    }

    private fun escapeJson(text: String): String {
        return text
            .replace("\\", "\\\\")
            .replace("\"", "\\\"")
            .replace("\n", "\\n")
            .replace("\r", "\\r")
    }
}
