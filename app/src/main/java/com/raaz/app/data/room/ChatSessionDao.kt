package com.raaz.app.data.room

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query

@Dao
interface ChatSessionDao {
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertSession(session: ChatSession)

    @Query("DELETE FROM chat_sessions WHERE startedAt < :cutoffMs")
    suspend fun deleteSessionsOlderThan(cutoffMs: Long)

    @Query("SELECT * FROM chat_sessions WHERE sessionId = :sessionId LIMIT 1")
    suspend fun getSession(sessionId: String): ChatSession?

    @Query("UPDATE chat_sessions SET partnerAlias = :alias WHERE sessionId = :sessionId")
    suspend fun updatePartnerAlias(sessionId: String, alias: String)

    @Query("UPDATE chat_sessions SET extensionsUsed = extensionsUsed + 1 WHERE sessionId = :sessionId")
    suspend fun incrementExtensionsUsed(sessionId: String)
}
