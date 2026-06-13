package com.raaz.app.features.chat

import android.app.Application
import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.work.ExistingWorkPolicy
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import com.raaz.app.data.repository.ChatRepository
import com.raaz.app.data.repository.VaultRepository
import com.raaz.app.data.room.RaazDatabase
import com.raaz.app.data.websocket.WebSocketClient
import com.raaz.app.data.worker.DeletionWorker
import java.util.UUID
import java.util.concurrent.TimeUnit

// H-8: constructor takes pre-loaded DataStore values — no runBlocking on main thread.
// See ChatScreen for the async loading logic (LaunchedEffect + OnboardingDataStore).
class ChatViewModelFactory(
    private val application: Application,
    private val wsBaseUrl: String,
    private val anonymousId: String,
    private val userAlias: String,
    private val ageBucket: String,
    private val city: String,
    private val prompt: String = ""
) : ViewModelProvider.Factory {

    @Suppress("UNCHECKED_CAST")
    override fun <T : ViewModel> create(modelClass: Class<T>): T {
        if (modelClass.isAssignableFrom(ChatViewModel::class.java)) {
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

            scheduleDeletion(application, sessionStart)

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

    companion object {
        // M-12: schedule an exact OneTimeWorkRequest at sessionStart + 48 h instead
        // of a periodic 12 h worker that over-deletes and has imprecise timing.
        fun scheduleDeletion(application: Application, sessionStartMs: Long) {
            val delayMs = (sessionStartMs + 48L * 3_600_000) - System.currentTimeMillis()
            if (delayMs <= 0) return
            val request = OneTimeWorkRequestBuilder<DeletionWorker>()
                .setInitialDelay(delayMs, TimeUnit.MILLISECONDS)
                .build()
            WorkManager.getInstance(application).enqueueUniqueWork(
                "RaazDeletion_$sessionStartMs",
                ExistingWorkPolicy.REPLACE,
                request
            )
        }
    }
}
