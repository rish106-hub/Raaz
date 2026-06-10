package com.raaz.app.features.vault

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.raaz.app.data.repository.VaultRepository
import com.raaz.app.data.room.RaazDatabase
import com.raaz.app.data.room.VaultMessage
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

class VaultViewModel(application: Application) : AndroidViewModel(application) {

    private val vaultRepository: VaultRepository = VaultRepository(
        RaazDatabase.getInstance(application).vaultDao()
    )

    val messages: StateFlow<List<VaultMessage>> = vaultRepository
        .getAllMessages()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    fun deleteMessage(id: Long) {
        viewModelScope.launch {
            vaultRepository.deleteMessage(id)
        }
    }
}
