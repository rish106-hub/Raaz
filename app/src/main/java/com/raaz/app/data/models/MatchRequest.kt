package com.raaz.app.data.models

data class MatchRequest(
    val promptId: String,
    val ageBucket: String,
    val city: String,
    val anonymousId: String,
    val timestamp: Long = System.currentTimeMillis()
)
