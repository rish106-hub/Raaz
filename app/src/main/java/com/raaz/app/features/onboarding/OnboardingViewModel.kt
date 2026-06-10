package com.raaz.app.features.onboarding

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.raaz.app.data.OnboardingDataStore
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class OnboardingState(
    val page: Int = 0,
    val selectedCategory: String = "",
    val selectedAgeBracket: String = "",
    val selectedCity: String = ""
)

class OnboardingViewModel(application: Application) : AndroidViewModel(application) {
    private val _state = MutableStateFlow(OnboardingState())
    val state: StateFlow<OnboardingState> = _state.asStateFlow()

    fun selectCategory(category: String) {
        _state.value = _state.value.copy(selectedCategory = category)
    }

    fun selectAgeBracket(bracket: String) {
        _state.value = _state.value.copy(selectedAgeBracket = bracket)
    }

    fun selectCity(city: String) {
        _state.value = _state.value.copy(selectedCity = city)
    }

    fun nextPage() {
        _state.value = _state.value.copy(page = _state.value.page + 1)
    }

    fun complete(onDone: () -> Unit) {
        viewModelScope.launch {
            val s = _state.value
            OnboardingDataStore.saveOnboarding(
                context = getApplication(),
                category = s.selectedCategory,
                ageBracket = s.selectedAgeBracket,
                city = s.selectedCity
            )
            onDone()
        }
    }
}
