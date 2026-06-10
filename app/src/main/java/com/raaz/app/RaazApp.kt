package com.raaz.app

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.navigation.compose.rememberNavController
import com.raaz.app.data.OnboardingDataStore
import com.raaz.app.navigation.NavGraph
import com.raaz.app.navigation.Screen
import com.raaz.app.ui.theme.RaazBackground

@Composable
fun RaazApp() {
    val context = LocalContext.current
    val navController = rememberNavController()
    val isOnboardingComplete by OnboardingDataStore.isOnboardingComplete(context)
        .collectAsState(initial = null)

    if (isOnboardingComplete == null) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(RaazBackground)
        )
        return
    }

    val startDestination = if (isOnboardingComplete == true) Screen.Home.route else Screen.Onboarding.route
    NavGraph(navController = navController, startDestination = startDestination)
}
