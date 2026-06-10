package com.raaz.app.data.repository

import com.raaz.app.data.room.VaultDao
import com.raaz.app.data.room.VaultMessage
import kotlinx.coroutines.flow.Flow

class VaultRepository(private val vaultDao: VaultDao) {
    fun getAllMessages(): Flow<List<VaultMessage>> = vaultDao.getAllMessages()

    suspend fun saveMessage(message: VaultMessage) = vaultDao.insertMessage(message)

    suspend fun deleteMessage(id: Long) = vaultDao.deleteMessage(id)

    suspend fun isMessageSaved(messageId: String): Boolean = vaultDao.isMessageSaved(messageId)
}
