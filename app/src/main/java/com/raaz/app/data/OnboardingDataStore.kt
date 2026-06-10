package com.raaz.app.data

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.map
import java.util.UUID

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "onboarding")

object OnboardingDataStore {
    private val KEY_COMPLETE = booleanPreferencesKey("onboarding_complete")
    private val KEY_CATEGORY = stringPreferencesKey("preferred_category")
    private val KEY_AGE_BRACKET = stringPreferencesKey("age_bracket")
    private val KEY_CITY = stringPreferencesKey("city")
    private val KEY_SESSION_ID = stringPreferencesKey("session_id")
    private val KEY_SESSION_ALIAS = stringPreferencesKey("session_alias")

    fun isOnboardingComplete(context: Context): Flow<Boolean> =
        context.dataStore.data.map { it[KEY_COMPLETE] ?: false }

    suspend fun saveOnboarding(
        context: Context,
        category: String,
        ageBracket: String,
        city: String
    ) {
        context.dataStore.edit { prefs ->
            prefs[KEY_CATEGORY] = category
            prefs[KEY_AGE_BRACKET] = ageBracket
            prefs[KEY_CITY] = city
            prefs[KEY_COMPLETE] = true
        }
    }

    fun getAnonymousAuth(context: Context): Flow<Pair<String, String>> =
        context.dataStore.data.map { prefs ->
            val sessionId = prefs[KEY_SESSION_ID] ?: ""
            val sessionAlias = prefs[KEY_SESSION_ALIAS] ?: ""
            Pair(sessionId, sessionAlias)
        }

    suspend fun generateAndSaveAnonymousAuth(context: Context): Pair<String, String> {
        val currentData = context.dataStore.data.map { prefs ->
            Pair(prefs[KEY_SESSION_ID] ?: "", prefs[KEY_SESSION_ALIAS] ?: "")
        }

        var sessionId: String = ""
        var sessionAlias: String = ""
        currentData.collect { (id, alias) ->
            sessionId = id
            sessionAlias = alias
        }

        if (sessionId.isEmpty()) {
            sessionId = UUID.randomUUID().toString()
        }

        if (sessionAlias.isEmpty()) {
            val randomNumber = (1000..9999).random()
            sessionAlias = "Raaz #$randomNumber"
        }

        context.dataStore.edit { prefs ->
            prefs[KEY_SESSION_ID] = sessionId
            prefs[KEY_SESSION_ALIAS] = sessionAlias
        }

        return Pair(sessionId, sessionAlias)
    }
}
