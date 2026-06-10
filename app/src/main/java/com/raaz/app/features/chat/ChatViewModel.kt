package com.raaz.app.features.chat

import android.os.CountDownTimer
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import android.app.Application
import com.raaz.app.data.models.ChatMessage
import com.raaz.app.data.repository.ChatRepository
import com.raaz.app.data.repository.ExtensionRequest
import com.raaz.app.data.repository.ExtensionResponse
import com.raaz.app.data.repository.HandleExchangeState
import com.raaz.app.data.repository.HandleReveal
import com.raaz.app.data.repository.TypingState
import com.raaz.app.data.websocket.WebSocketClient
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class ChatViewModel(
    application: Application,
    private val chatRepository: ChatRepository? = null,
    private val currentUserAlias: String = "Raaz #${(1000..9999).random()}"
) : AndroidViewModel(application) {
    private val _timerText = MutableStateFlow("20:00")
    val timerText: StateFlow<String> = _timerText.asStateFlow()

    private val _inputText = MutableStateFlow("")
    val inputText: StateFlow<String> = _inputText.asStateFlow()

    private val _sessionAlias = MutableStateFlow(currentUserAlias)
    val sessionAlias: StateFlow<String> = _sessionAlias.asStateFlow()

    private val _messages = MutableStateFlow<List<ChatMessage>>(emptyList())
    val messages: StateFlow<List<ChatMessage>> = _messages.asStateFlow()

    private val _partnerTyping = MutableStateFlow(false)
    val partnerTyping: StateFlow<Boolean> = _partnerTyping.asStateFlow()

    private val _isTyping = MutableStateFlow(false)
    val isTyping: StateFlow<Boolean> = _isTyping.asStateFlow()

    private val _extensionRequest = MutableStateFlow<ExtensionRequest?>(null)
    val extensionRequest: StateFlow<ExtensionRequest?> = _extensionRequest.asStateFlow()

    private val _partnerHandleExchange = MutableStateFlow<HandleExchangeState?>(null)
    val partnerHandleExchange: StateFlow<HandleExchangeState?> = _partnerHandleExchange.asStateFlow()

    private val _userApprovedHandle = MutableStateFlow(false)
    val userApprovedHandle: StateFlow<Boolean> = _userApprovedHandle.asStateFlow()

    private val _partnerHandle = MutableStateFlow<String?>(null)
    val partnerHandle: StateFlow<String?> = _partnerHandle.asStateFlow()

    private var countDownTimer: CountDownTimer? = null
    private var typingDebounceJob: Job? = null
    private var typingStopJob: Job? = null
    private var timerMs: Long = 20 * 60 * 1000L
    private val userHandle: String = "Raaz/${(10000..99999).random()}"

    init {
        startTimer()
        subscribeToMessages()
        subscribeToTyping()
        subscribeToExtensionEvents()
        subscribeToHandleExchange()
    }

    private fun subscribeToMessages() {
        if (chatRepository != null) {
            viewModelScope.launch {
                chatRepository.getMessages().collect { message ->
                    _messages.value = _messages.value + message
                }
            }
        }
    }

    private fun subscribeToTyping() {
        if (chatRepository != null) {
            viewModelScope.launch {
                chatRepository.getTypingState().collect { state ->
                    _partnerTyping.value = state.isTyping
                }
            }
        }
    }

    private fun subscribeToExtensionEvents() {
        if (chatRepository != null) {
            viewModelScope.launch {
                chatRepository.getExtensionRequests().collect { request ->
                    _extensionRequest.value = request
                }
            }
            viewModelScope.launch {
                chatRepository.getExtensionResponses().collect { response ->
                    if (response.approved) {
                        timerMs += 600 * 1000
                    }
                    _extensionRequest.value = null
                }
            }
        }
    }

    private fun subscribeToHandleExchange() {
        if (chatRepository != null) {
            viewModelScope.launch {
                chatRepository.getHandleExchangeEvents().collect { exchange ->
                    _partnerHandleExchange.value = exchange
                }
            }
            viewModelScope.launch {
                chatRepository.getHandleReveals().collect { reveal ->
                    _partnerHandle.value = reveal.partnerHandle
                }
            }
        }
    }

    private fun startTimer() {
        countDownTimer = object : CountDownTimer(20 * 60 * 1000L, 1000) {
            override fun onTick(millisUntilFinished: Long) {
                val minutes = millisUntilFinished / 60000
                val seconds = (millisUntilFinished % 60000) / 1000
                _timerText.value = "%02d:%02d".format(minutes, seconds)
            }
            override fun onFinish() {
                _timerText.value = "00:00"
            }
        }.start()
    }

    fun onInputChange(text: String) {
        _inputText.value = text

        typingStopJob?.cancel()
        typingDebounceJob?.cancel()

        if (text.isNotEmpty() && !_isTyping.value) {
            typingDebounceJob = viewModelScope.launch {
                delay(300)
                _isTyping.value = true
                chatRepository?.sendTyping(true)
            }
        }

        if (text.isEmpty() && _isTyping.value) {
            _isTyping.value = false
            chatRepository?.sendTyping(false)
        } else if (text.isNotEmpty()) {
            typingStopJob = viewModelScope.launch {
                delay(1000)
                if (_inputText.value.isNotEmpty()) {
                    _isTyping.value = false
                    chatRepository?.sendTyping(false)
                }
            }
        }
    }

    fun sendMessage() {
        val text = _inputText.value.trim()
        if (text.isNotEmpty() && chatRepository != null) {
            chatRepository.sendMessage(text)
            _inputText.value = ""
        }
    }

    fun requestExtension(userId: String) {
        chatRepository?.sendExtensionRequest(userId)
    }

    fun respondToExtension(approved: Boolean) {
        chatRepository?.sendExtensionResponse(approved)
        _extensionRequest.value = null
    }

    fun initiateHandleExchange(userId: String) {
        chatRepository?.sendHandleExchange(true, userId)
        _userApprovedHandle.value = true
    }

    fun respondToHandleExchange(approved: Boolean) {
        chatRepository?.sendHandleExchange(approved, partnerHandleExchange.value?.userId ?: "")
        if (approved) {
            _userApprovedHandle.value = true
        }
        _partnerHandleExchange.value = null
    }

    fun getVisibleUserHandle(): String? {
        return if (_userApprovedHandle.value && _partnerHandle.value != null) userHandle else null
    }

    private fun formatTime(milliseconds: Long): String {
        val minutes = milliseconds / 60000
        val seconds = (milliseconds % 60000) / 1000
        return "%02d:%02d".format(minutes, seconds)
    }

    override fun onCleared() {
        countDownTimer?.cancel()
        typingDebounceJob?.cancel()
        typingStopJob?.cancel()
        chatRepository?.disconnect()
        super.onCleared()
    }
}
