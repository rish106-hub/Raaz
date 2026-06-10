package com.raaz.app.navigation

import android.net.Uri
import androidx.compose.runtime.Composable
import androidx.navigation.NavHostController
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.raaz.app.features.chat.ChatScreen
import com.raaz.app.features.home.HomeScreen
import com.raaz.app.features.onboarding.OnboardingScreen
import com.raaz.app.features.vault.VaultScreen

@Composable
fun NavGraph(
    navController: NavHostController,
    startDestination: String
) {
    NavHost(navController = navController, startDestination = startDestination) {
        composable(Screen.Onboarding.route) {
            OnboardingScreen(
                onComplete = {
                    navController.navigate(Screen.Home.route) {
                        popUpTo(Screen.Onboarding.route) { inclusive = true }
                    }
                }
            )
        }
        composable(Screen.Home.route) {
            HomeScreen(
                onEcho = { prompt ->
                    navController.navigate(Screen.Chat.createRoute(prompt))
                },
                onVault = {
                    navController.navigate(Screen.Vault.route)
                }
            )
        }
        composable(
            route = Screen.Chat.route,
            arguments = listOf(navArgument("prompt") { type = NavType.StringType })
        ) { backStackEntry ->
            val prompt = Uri.decode(backStackEntry.arguments?.getString("prompt").orEmpty())
            ChatScreen(
                prompt = prompt,
                onBack = { navController.popBackStack() }
            )
        }
        composable(Screen.Vault.route) {
            VaultScreen(
                onBack = { navController.popBackStack() }
            )
        }
    }
}
