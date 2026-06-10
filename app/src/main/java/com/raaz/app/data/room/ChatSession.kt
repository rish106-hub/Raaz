package com.raaz.app.data.room

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "chat_sessions")
data class ChatSession(
    @PrimaryKey val sessionId: String,
    val startedAt: Long,
    val prompt: String,
    val partnerAlias: String
)
