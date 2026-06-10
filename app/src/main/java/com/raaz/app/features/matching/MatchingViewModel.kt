package com.raaz.app.features.matching

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.raaz.app.data.OnboardingDataStore
import com.raaz.app.data.models.MatchRequest
import com.raaz.app.data.repository.AuthRepository
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.launch

class MatchingViewModel(
    private val authRepository: AuthRepository,
    application: Application
) : AndroidViewModel(application) {

    private val _matchingState = MutableStateFlow<MatchingState>(MatchingState.Idle)
    val matchingState: StateFlow<MatchingState> = _matchingState.asStateFlow()

    private var timerJob: Job? = null
    private val TIMEOUT_SECONDS = 900L // 15 minutes

    fun startMatching(promptId: String) {
        viewModelScope.launch {
            val request = buildMatchRequest(promptId)
            if (!canRequestMatch(request)) {
                _matchingState.value = MatchingState.Error("Missing required fields for matching")
                return@launch
            }

            _matchingState.value = MatchingState.InQueue(0)
            startTimer()
        }
    }

    private suspend fun buildMatchRequest(promptId: String): MatchRequest {
        val context = getApplication<Application>()
        var ageBucket = ""
        var city = ""
        var anonymousId = ""

        OnboardingDataStore.getAnonymousAuth(context).collect { (sessionId, _) ->
            anonymousId = sessionId
        }

        OnboardingDataStore.getAgeBracket(context).collect { bracket ->
            ageBucket = bracket
        }

        OnboardingDataStore.getCity(context).collect { c ->
            city = c
        }

        return MatchRequest(
            promptId = promptId,
            ageBucket = ageBucket,
            city = city,
            anonymousId = anonymousId
        )
    }

    fun canRequestMatch(request: MatchRequest): Boolean {
        return request.promptId.isNotEmpty() &&
                request.ageBucket.isNotEmpty() &&
                request.city.isNotEmpty() &&
                request.anonymousId.isNotEmpty()
    }

    private fun startTimer() {
        timerJob = viewModelScope.launch {
            val timerFlow = flow {
                repeat(Int.MAX_VALUE) {
                    emit(it.toLong())
                    delay(1000)
                }
            }

            timerFlow.collect { elapsedSeconds ->
                if (elapsedSeconds >= TIMEOUT_SECONDS) {
                    _matchingState.value = MatchingState.TimeoutFallback(elapsedSeconds)
                } else {
                    _matchingState.value = MatchingState.InQueue(elapsedSeconds)
                }
            }
        }
    }

    fun cancelMatching() {
        timerJob?.cancel()
        _matchingState.value = MatchingState.Idle
    }

    override fun onCleared() {
        timerJob?.cancel()
        super.onCleared()
    }
}
