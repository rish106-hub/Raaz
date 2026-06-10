package com.raaz.app.features.home

import androidx.lifecycle.ViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import java.time.LocalDate

private val PROMPTS = listOf(
    "What will your parents never fully understand about you?",
    "What have you given up on that still hurts?",
    "What do you want, but are too scared to admit?"
)

class HomeViewModel : ViewModel() {
    private val _todayPrompt = MutableStateFlow(todayPrompt())
    val todayPrompt: StateFlow<String> = _todayPrompt.asStateFlow()

    private fun todayPrompt(): String {
        val dayIndex = LocalDate.now().dayOfYear
        return PROMPTS[dayIndex % PROMPTS.size]
    }
}
