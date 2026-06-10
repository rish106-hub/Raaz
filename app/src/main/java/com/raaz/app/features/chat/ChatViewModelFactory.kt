package com.raaz.app.features.chat

import android.app.Application
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import com.raaz.app.data.repository.ChatRepository
import com.raaz.app.data.websocket.WebSocketClient

class ChatViewModelFactory(
    private val application: Application,
    private val wsUrl: String,
    private val userAlias: String
) : ViewModelProvider.Factory {

    @Suppress("UNCHECKED_CAST")
    override fun <T : ViewModel> create(modelClass: Class<T>): T {
        if (modelClass.isAssignableFrom(ChatViewModel::class.java)) {
            val wsClient = WebSocketClient(wsUrl).also { it.connect() }
            val repo = ChatRepository(wsClient, userAlias)
            return ChatViewModel(application, repo, userAlias) as T
        }
        throw IllegalArgumentException("Unknown ViewModel class: ${modelClass.name}")
    }
}
