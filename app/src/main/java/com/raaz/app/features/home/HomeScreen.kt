package com.raaz.app.features.home

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.systemBarsPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.raaz.app.data.repository.AuthRepositoryImpl
import com.raaz.app.features.matching.MatchingOverlay
import com.raaz.app.features.matching.MatchingViewModel
import com.raaz.app.features.onboarding.RaazButton
import com.raaz.app.ui.theme.RaazAccent
import com.raaz.app.ui.theme.RaazBackground
import com.raaz.app.ui.theme.RaazSurface
import java.time.LocalDate
import java.time.format.DateTimeFormatter
import java.util.Locale

@Composable
fun HomeScreen(
    onEcho: (String) -> Unit,
    viewModel: HomeViewModel = viewModel(),
    matchingViewModel: MatchingViewModel? = null
) {
    val context = androidx.compose.ui.platform.LocalContext.current
    val application = context.applicationContext as android.app.Application
    val authRepository = remember { AuthRepositoryImpl(context) }
    val actualMatchingViewModel = matchingViewModel ?: viewModel {
        MatchingViewModel(authRepository, application)
    }

    val prompt by viewModel.todayPrompt.collectAsState()
    val matchingState by actualMatchingViewModel.matchingState.collectAsState()

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(RaazBackground)
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .systemBarsPadding()
                .padding(24.dp),
            verticalArrangement = Arrangement.SpaceBetween
        ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = "Raaz",
                color = Color.White,
                fontSize = 28.sp,
                fontWeight = FontWeight.Bold,
                letterSpacing = (-0.5).sp
            )
            Text(
                text = LocalDate.now()
                    .format(DateTimeFormatter.ofPattern("d MMM", Locale.ENGLISH)),
                color = Color.White.copy(alpha = 0.35f),
                fontSize = 14.sp
            )
        }

        Column {
            Text(
                text = "TODAY'S PROMPT",
                color = RaazAccent,
                fontSize = 11.sp,
                fontWeight = FontWeight.SemiBold,
                letterSpacing = 2.sp
            )
            Spacer(modifier = Modifier.height(12.dp))
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(RaazSurface, RoundedCornerShape(12.dp))
                    .border(1.dp, Color.White.copy(alpha = 0.07f), RoundedCornerShape(12.dp))
                    .padding(24.dp)
            ) {
                Text(
                    text = prompt,
                    color = Color.White,
                    fontSize = 22.sp,
                    fontWeight = FontWeight.Medium,
                    lineHeight = 32.sp
                )
            }
        }

        Column {
            Text(
                text = "Tap Echo to get matched with someone\nwho chose the same prompt.",
                color = Color.White.copy(alpha = 0.35f),
                fontSize = 13.sp,
                lineHeight = 20.sp
            )
            Spacer(modifier = Modifier.height(12.dp))
            RaazButton(text = "Echo", onClick = { actualMatchingViewModel.startMatching(prompt) })
        }
        } // end inner Column

        MatchingOverlay(
            state = matchingState,
            onCancel = { actualMatchingViewModel.cancelMatching() },
            onMatched = {
                actualMatchingViewModel.cancelMatching()
                onEcho(prompt)
            }
        )
    } // end Box
}
