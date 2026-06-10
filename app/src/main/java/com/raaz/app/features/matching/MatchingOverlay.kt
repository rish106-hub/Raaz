package com.raaz.app.features.matching

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.raaz.app.features.chat.TimerComponent
import com.raaz.app.features.onboarding.RaazButton
import com.raaz.app.ui.theme.RaazBackground

@Composable
fun MatchingOverlay(
    state: MatchingState,
    onCancel: () -> Unit,
    modifier: Modifier = Modifier
) {
    if (state is MatchingState.Idle) {
        return
    }

    Box(
        modifier = modifier
            .fillMaxSize()
            .background(RaazBackground.copy(alpha = 0.95f)),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            when (state) {
                is MatchingState.InQueue -> {
                    val formattedTime = formatSeconds(state.elapsedSeconds)
                    Text(
                        text = "Finding your match...",
                        color = Color.White,
                        fontSize = 24.sp,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(modifier = Modifier.height(24.dp))
                    TimerComponent(timerText = formattedTime)
                    Spacer(modifier = Modifier.height(24.dp))
                    Text(
                        text = "Time elapsed: ${formatSeconds(state.elapsedSeconds)}",
                        color = Color.White.copy(alpha = 0.6f),
                        fontSize = 14.sp
                    )
                }

                is MatchingState.TimeoutFallback -> {
                    val formattedTime = formatSeconds(state.elapsedSeconds)
                    Text(
                        text = "National pool activated",
                        color = Color.White,
                        fontSize = 24.sp,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(modifier = Modifier.height(12.dp))
                    Text(
                        text = "Searching nationwide",
                        color = Color.White.copy(alpha = 0.6f),
                        fontSize = 14.sp
                    )
                    Spacer(modifier = Modifier.height(24.dp))
                    TimerComponent(timerText = formattedTime)
                }

                is MatchingState.Matched -> {
                    Text(
                        text = "Connected!",
                        color = Color.White,
                        fontSize = 24.sp,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(modifier = Modifier.height(12.dp))
                    Text(
                        text = "Matched with ${state.partnerAlias}",
                        color = Color.White.copy(alpha = 0.6f),
                        fontSize = 14.sp
                    )
                }

                is MatchingState.Error -> {
                    Text(
                        text = "Error",
                        color = Color.White,
                        fontSize = 24.sp,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(modifier = Modifier.height(12.dp))
                    Text(
                        text = state.message,
                        color = Color.White.copy(alpha = 0.6f),
                        fontSize = 14.sp
                    )
                }

                is MatchingState.Idle -> {
                    // Already handled at start of function
                }
            }

            Spacer(modifier = Modifier.height(32.dp))
            RaazButton(text = "Cancel", onClick = onCancel)
        }
    }
}

private fun formatSeconds(seconds: Long): String {
    val minutes = seconds / 60
    val secs = seconds % 60
    return "%02d:%02d".format(minutes, secs)
}
