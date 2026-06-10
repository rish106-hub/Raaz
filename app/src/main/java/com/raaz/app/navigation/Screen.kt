package com.raaz.app.navigation

import android.net.Uri

sealed class Screen(val route: String) {
    object Onboarding : Screen("onboarding")
    object Home : Screen("home")
    object Chat : Screen("chat/{prompt}") {
        fun createRoute(prompt: String) = "chat/${Uri.encode(prompt)}"
    }
    object Vault : Screen("vault")
}
