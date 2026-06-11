package com.raaz.app.analytics

import android.content.Context
import com.google.firebase.analytics.FirebaseAnalytics

/**
 * Singleton analytics wrapper. All events are non-PII — only event names,
 * no user identifiers (DPDP Act 2023 compliant).
 *
 * Call init(context) once from RaazApplication.onCreate().
 */
object RaazAnalytics {
    private var analytics: FirebaseAnalytics? = null

    fun init(context: Context) {
        analytics = FirebaseAnalytics.getInstance(context)
    }

    fun matchInitiated() = log("match_initiated")
    fun matchSuccess() = log("match_success")
    fun sessionExtended() = log("session_extended")
    fun handleExchanged() = log("handle_exchanged")
    fun vaultSave() = log("vault_save")
    fun crisisTriggerSeen() = log("crisis_trigger_seen")

    private fun log(event: String) {
        analytics?.logEvent(event, null)
    }
}
