package com.raaz.app.data.repository

import android.content.Context
import com.raaz.app.data.OnboardingDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

class AuthRepositoryImpl(private val context: Context) : AuthRepository {
    override fun getAnonymousAuth(): Flow<AuthRepository.AuthState> {
        return OnboardingDataStore.getAnonymousAuth(context)
            .map { (sessionId, sessionAlias) ->
                AuthRepository.AuthState(
                    sessionId = sessionId,
                    sessionAlias = sessionAlias
                )
            }
    }
}
