package com.raaz.app

import android.app.Application
import android.app.NotificationChannel
import android.app.NotificationManager
import android.os.Build
import com.google.firebase.crashlytics.FirebaseCrashlytics
import com.raaz.app.analytics.RaazAnalytics

class RaazApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        FirebaseCrashlytics.getInstance()
            .setCrashlyticsCollectionEnabled(!BuildConfig.DEBUG)
        RaazAnalytics.init(this)
        createNotificationChannels()
    }

    // H-11: channel must exist before any FCM notification is posted on API 26+
    private fun createNotificationChannels() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                "Raaz",
                NotificationManager.IMPORTANCE_DEFAULT
            ).apply {
                description = "Raaz match and session notifications"
            }
            getSystemService(NotificationManager::class.java)
                .createNotificationChannel(channel)
        }
    }

    companion object {
        const val CHANNEL_ID = "raaz_channel"
    }
}
