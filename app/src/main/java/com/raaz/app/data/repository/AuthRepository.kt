package com.raaz.app.data.repository

import kotlinx.coroutines.flow.Flow

interface AuthRepository {
    fun getAnonymousAuth(): Flow<AuthState>

    data class AuthState(
        val sessionId: String,
        val sessionAlias: String
    )
}
