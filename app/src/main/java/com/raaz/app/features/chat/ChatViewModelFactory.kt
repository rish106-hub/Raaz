package com.raaz.app.features.chat

import android.app.Application
import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import com.raaz.app.data.OnboardingDataStore
import com.raaz.app.data.repository.ChatRepository
import com.raaz.app.data.repository.VaultRepository
import com.raaz.app.data.room.RaazDatabase
import com.raaz.app.data.websocket.WebSocketClient
import com.raaz.app.data.worker.DeletionWorker
import java.util.UUID
import java.util.concurrent.TimeUnit
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking

class ChatViewModelFactory(
    private val application: Application,
    private val wsBaseUrl: String,
    private val prompt: String = ""
) : ViewModelProvider.Factory {

    @Suppress("UNCHECKED_CAST")
    override fun <T : ViewModel> create(modelClass: Class<T>): T {
        if (modelClass.isAssignableFrom(ChatViewModel::class.java)) {
            // Read registration params synchronously. DataStore serves from an
            // in-memory cache after the first read, so this is fast (< 5 ms).
            val authPair = runBlocking { OnboardingDataStore.getAnonymousAuth(application).first() }
            val ageBucket = runBlocking { OnboardingDataStore.getAgeBracket(application).first() }
            val city = runBlocking { OnboardingDataStore.getCity(application).first() }

            val anonymousId = authPair.first.ifEmpty { UUID.randomUUID().toString() }
            val userAlias = authPair.second.ifEmpty { "Raaz #${(1000..9999).random()}" }

            val wsUrl = Uri.parse(wsBaseUrl)
                .buildUpon()
                .appendQueryParameter("promptId", prompt)
                .appendQueryParameter("ageBucket", ageBucket)
                .appendQueryParameter("city", city)
                .appendQueryParameter("anonymousId", anonymousId)
                .build()
                .toString()

            val db = RaazDatabase.getInstance(application)
            val wsClient = WebSocketClient(wsUrl).also { it.connect() }
            val chatRepo = ChatRepository(wsClient, userAlias)
            val vaultRepo = VaultRepository(db.vaultDao())
            val sessionId = UUID.randomUUID().toString()
            val sessionStart = System.currentTimeMillis()

            scheduleDeletion()

            return ChatViewModel(
                application = application,
                chatRepository = chatRepo,
                currentUserAlias = userAlias,
                vaultRepository = vaultRepo,
                chatSessionDao = db.chatSessionDao(),
                sessionId = sessionId,
                sessionStartTime = sessionStart,
                prompt = prompt
            ) as T
        }
        throw IllegalArgumentException("Unknown ViewModel class: ${modelClass.name}")
    }

    private fun scheduleDeletion() {
        WorkManager.getInstance(application).enqueueUniquePeriodicWork(
            "RaazEphemeralDeletion",
            ExistingPeriodicWorkPolicy.KEEP,
            PeriodicWorkRequestBuilder<DeletionWorker>(12, TimeUnit.HOURS).build()
        )
    }
}
