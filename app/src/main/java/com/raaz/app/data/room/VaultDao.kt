package com.raaz.app.data.room

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import kotlinx.coroutines.flow.Flow

@Dao
interface VaultDao {
    @Insert(onConflict = OnConflictStrategy.IGNORE)
    suspend fun insertMessage(message: VaultMessage)

    @Query("SELECT * FROM vault_messages ORDER BY savedAt DESC")
    fun getAllMessages(): Flow<List<VaultMessage>>

    @Query("DELETE FROM vault_messages WHERE id = :id")
    suspend fun deleteMessage(id: Long)

    @Query("SELECT EXISTS(SELECT 1 FROM vault_messages WHERE messageId = :messageId)")
    suspend fun isMessageSaved(messageId: String): Boolean
}
