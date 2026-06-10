package com.raaz.app.data.models

data class ChatMessage(
    val messageId: String,
    val text: String,
    val senderAlias: String,
    val isOwnMessage: Boolean,
    val timestamp: Long
)
