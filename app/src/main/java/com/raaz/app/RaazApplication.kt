package com.raaz.app

import android.app.Application
import com.google.firebase.crashlytics.FirebaseCrashlytics
import com.raaz.app.analytics.RaazAnalytics

class RaazApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        // Disable Crashlytics in debug builds — keeps the Firebase dashboard clean.
        FirebaseCrashlytics.getInstance()
            .setCrashlyticsCollectionEnabled(!BuildConfig.DEBUG)
        RaazAnalytics.init(this)
    }
}
