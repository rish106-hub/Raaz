package com.raaz.app.features.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.raaz.app.BuildConfig
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import okhttp3.OkHttpClient
import okhttp3.Request
import org.json.JSONObject
import java.time.LocalDate
import java.util.concurrent.TimeUnit

private val FALLBACK_PROMPTS = listOf(
    "What will your parents never fully understand about you?",
    "What have you given up on that still hurts?",
    "What do you want, but are too scared to admit?",
    "What truth are you avoiding right now?",
    "Who do you wish knew you better?",
    "What would you do if you weren't afraid?",
    "What part of yourself do you hide most?",
    "What relationship still needs closure?",
    "When did you last feel truly seen?",
    "What do you keep pretending is fine?"
)

// H-9: fetch today's prompt from server; fall back to local list on any failure.
class HomeViewModel : ViewModel() {
    private val _todayPrompt = MutableStateFlow(localFallback())
    val todayPrompt: StateFlow<String> = _todayPrompt.asStateFlow()

    private val httpClient = OkHttpClient.Builder()
        .connectTimeout(5, TimeUnit.SECONDS)
        .readTimeout(5, TimeUnit.SECONDS)
        .build()

    init {
        fetchTodayPrompt()
    }

    private fun fetchTodayPrompt() {
        viewModelScope.launch {
            val fetched = withContext(Dispatchers.IO) {
                try {
                    val request = Request.Builder()
                        .url("${BuildConfig.API_BASE_URL}/prompts/today")
                        .get()
                        .build()
                    httpClient.newCall(request).execute().use { response ->
                        if (!response.isSuccessful) return@withContext null
                        val body = response.body?.string() ?: return@withContext null
                        JSONObject(body).optString("text").takeIf { it.isNotEmpty() }
                    }
                } catch (_: Exception) {
                    null
                }
            }
            if (fetched != null) {
                _todayPrompt.value = fetched
            }
        }
    }

    override fun onCleared() {
        super.onCleared()
        httpClient.dispatcher.executorService.shutdown()
    }
}

private fun localFallback(): String {
    val dayIndex = LocalDate.now().dayOfYear
    return FALLBACK_PROMPTS[dayIndex % FALLBACK_PROMPTS.size]
}
