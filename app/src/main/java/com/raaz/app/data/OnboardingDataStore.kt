package com.raaz.app.data

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "onboarding")

object OnboardingDataStore {
    private val KEY_COMPLETE = booleanPreferencesKey("onboarding_complete")
    private val KEY_CATEGORY = stringPreferencesKey("preferred_category")
    private val KEY_AGE_BRACKET = stringPreferencesKey("age_bracket")
    private val KEY_CITY = stringPreferencesKey("city")

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
}
