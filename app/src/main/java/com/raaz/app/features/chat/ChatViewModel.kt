package com.raaz.app.features.chat

import android.os.CountDownTimer
import androidx.lifecycle.ViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class ChatViewModel : ViewModel() {
    private val _timerText = MutableStateFlow("20:00")
    val timerText: StateFlow<String> = _timerText.asStateFlow()

    private val _inputText = MutableStateFlow("")
    val inputText: StateFlow<String> = _inputText.asStateFlow()

    private val _sessionAlias = MutableStateFlow("Raaz #${(1000..9999).random()}")
    val sessionAlias: StateFlow<String> = _sessionAlias.asStateFlow()

    private var countDownTimer: CountDownTimer? = null

    init {
        startTimer()
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
    }

    fun sendMessage() {
        _inputText.value = ""
    }

    override fun onCleared() {
        countDownTimer?.cancel()
        super.onCleared()
    }
}
