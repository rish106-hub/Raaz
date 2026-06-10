package com.raaz.app.data.room

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "vault_messages")
data class VaultMessage(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val messageId: String,
    val text: String,
    val senderAlias: String,
    val sessionId: String,
    val prompt: String,
    val savedAt: Long
)
