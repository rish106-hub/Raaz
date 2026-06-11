package com.raaz.app.features.chat

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.raaz.app.data.models.ChatMessage
import com.raaz.app.data.repository.ChatRepository
import com.raaz.app.data.repository.CrisisTriggerData
import com.raaz.app.data.repository.ExtensionRequest
import com.raaz.app.data.repository.ModerationAlertData
import com.raaz.app.data.repository.VaultRepository
import com.raaz.app.data.room.ChatSession
import com.raaz.app.data.room.ChatSessionDao
import com.raaz.app.data.room.VaultMessage
import com.raaz.app.data.websocket.ConnectionState
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

class ChatViewModel(
    application: Application,
    private val chatRepository: ChatRepository? = null,
    private val currentUserAlias: String = "Raaz #${(1000..9999).random()}",
    private val vaultRepository: VaultRepository? = null,
    private val chatSessionDao: ChatSessionDao? = null,
    val sessionId: String = java.util.UUID.randomUUID().toString(),
    private val sessionStartTime: Long = System.currentTimeMillis(),
    private val prompt: String = ""
) : AndroidViewModel(application) {

    private val _timerText = MutableStateFlow("20:00")
    val timerText: StateFlow<String> = _timerText.asStateFlow()

    private val _deletionCountdown = MutableStateFlow("")
    val deletionCountdown: StateFlow<String> = _deletionCountdown.asStateFlow()

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

    private val _partnerHandleExchange = MutableStateFlow<com.raaz.app.data.repository.HandleExchangeState?>(null)
    val partnerHandleExchange: StateFlow<com.raaz.app.data.repository.HandleExchangeState?> = _partnerHandleExchange.asStateFlow()

    private val _userApprovedHandle = MutableStateFlow(false)
    val userApprovedHandle: StateFlow<Boolean> = _userApprovedHandle.asStateFlow()

    private val _partnerHandle = MutableStateFlow<String?>(null)
    val partnerHandle: StateFlow<String?> = _partnerHandle.asStateFlow()

    private val _pendingVaultMessage = MutableStateFlow<ChatMessage?>(null)
    val pendingVaultMessage: StateFlow<ChatMessage?> = _pendingVaultMessage.asStateFlow()

    private val _vaultSaveSuccess = MutableSharedFlow<Unit>(extraBufferCapacity = 1)
    val vaultSaveSuccess: SharedFlow<Unit> = _vaultSaveSuccess.asSharedFlow()

    private val _moderationAlert = MutableStateFlow<ModerationAlertData?>(null)
    val moderationAlert: StateFlow<ModerationAlertData?> = _moderationAlert.asStateFlow()

    private val _crisisTriggered = MutableStateFlow<CrisisTriggerData?>(null)
    val crisisTriggered: StateFlow<CrisisTriggerData?> = _crisisTriggered.asStateFlow()

    val connectionState: StateFlow<ConnectionState> = chatRepository
        ?.connectionState
        ?.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), ConnectionState.DISCONNECTED)
        ?: MutableStateFlow(ConnectionState.DISCONNECTED).asStateFlow()

    private var timerJob: Job? = null
    private var remainingMs: Long = 20 * 60 * 1000L
    private var typingDebounceJob: Job? = null
    private var typingStopJob: Job? = null
    private val userHandle: String = "Raaz/${(10000..99999).random()}"

    init {
        startTimer()
        startDeletionCountdown()
        recordSession()
        subscribeToMessages()
        subscribeToTyping()
        subscribeToExtensionEvents()
        subscribeToHandleExchange()
        subscribeToCrisisAndModeration()
    }

    private fun recordSession() {
        viewModelScope.launch {
            chatSessionDao?.insertSession(
                ChatSession(
                    sessionId = sessionId,
                    startedAt = sessionStartTime,
                    prompt = prompt,
                    partnerAlias = ""
                )
            )
        }
    }

    private fun startDeletionCountdown() {
        updateDeletionLabel()
        viewModelScope.launch {
            while (true) {
                delay(60_000L)
                updateDeletionLabel()
            }
        }
    }

    private fun updateDeletionLabel() {
        val deleteAt = sessionStartTime + 48L * 60 * 60 * 1000
        val remaining = deleteAt - System.currentTimeMillis()
        _deletionCountdown.value = if (remaining > 0) {
            val hours = remaining / (60 * 60 * 1000L)
            val minutes = (remaining % (60 * 60 * 1000L)) / (60 * 1000L)
            "Deletes in ${hours}h ${minutes}m"
        } else {
            "Data deleted"
        }
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
                        startTimer(remainingMs + 600_000L)
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

    private fun subscribeToCrisisAndModeration() {
        if (chatRepository != null) {
            viewModelScope.launch {
                chatRepository.getModerationAlerts().collect { alert ->
                    _moderationAlert.value = alert
                }
            }
            viewModelScope.launch {
                chatRepository.getCrisisTriggers().collect { crisis ->
                    _crisisTriggered.value = crisis
                }
            }
        }
    }

    fun dismissModerationAlert() {
        _moderationAlert.value = null
    }

    fun dismissCrisis() {
        _crisisTriggered.value = null
    }

    private fun startTimer(fromMs: Long = remainingMs) {
        timerJob?.cancel()
        remainingMs = fromMs
        timerJob = viewModelScope.launch {
            while (remainingMs > 0) {
                delay(1000)
                remainingMs -= 1000
                val minutes = remainingMs / 60000
                val seconds = (remainingMs % 60000) / 1000
                _timerText.value = "%02d:%02d".format(minutes, seconds)
            }
            _timerText.value = "00:00"
        }
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

    fun requestVaultSave(message: ChatMessage) {
        _pendingVaultMessage.value = message
    }

    fun confirmVaultSave() {
        val message = _pendingVaultMessage.value ?: return
        _pendingVaultMessage.value = null
        viewModelScope.launch {
            vaultRepository?.saveMessage(
                VaultMessage(
                    messageId = message.messageId,
                    text = message.text,
                    senderAlias = message.senderAlias,
                    sessionId = sessionId,
                    prompt = prompt,
                    savedAt = System.currentTimeMillis()
                )
            )
            _vaultSaveSuccess.emit(Unit)
        }
    }

    fun cancelVaultSave() {
        _pendingVaultMessage.value = null
    }

    override fun onCleared() {
        timerJob?.cancel()
        typingDebounceJob?.cancel()
        typingStopJob?.cancel()
        chatRepository?.disconnect()
        super.onCleared()
    }
}
