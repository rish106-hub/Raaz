package com.raaz.app.features.onboarding

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.raaz.app.ui.theme.RaazAccent
import com.raaz.app.ui.theme.RaazBackground

private val CATEGORIES = listOf("Identity", "Ambition", "Relationships", "Regret")
private val AGE_BRACKETS = listOf("18–22", "23–28")
private val CITIES = listOf("Bengaluru", "Mumbai", "Delhi NCR")

@Composable
fun OnboardingScreen(
    onComplete: () -> Unit,
    viewModel: OnboardingViewModel = viewModel()
) {
    val state by viewModel.state.collectAsState()

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(RaazBackground)
            .padding(top = 48.dp, start = 24.dp, end = 24.dp, bottom = 32.dp)
    ) {
        when (state.page) {
            0 -> IntroPage(onNext = { viewModel.nextPage() })
            1 -> CategoryPage(
                selectedCategory = state.selectedCategory,
                onSelect = viewModel::selectCategory,
                onNext = { if (state.selectedCategory.isNotEmpty()) viewModel.nextPage() }
            )
            2 -> DemographicsPage(
                selectedAgeBracket = state.selectedAgeBracket,
                selectedCity = state.selectedCity,
                onSelectAge = viewModel::selectAgeBracket,
                onSelectCity = viewModel::selectCity,
                onComplete = {
                    if (state.selectedAgeBracket.isNotEmpty() && state.selectedCity.isNotEmpty()) {
                        viewModel.complete(onComplete)
                    }
                }
            )
        }
    }
}

@Composable
private fun IntroPage(onNext: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize(),
        verticalArrangement = Arrangement.SpaceBetween
    ) {
        Column {
            Text(
                text = "Raaz",
                color = Color.White,
                fontSize = 40.sp,
                fontWeight = FontWeight.Bold,
                letterSpacing = (-1).sp
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = "Say what you\nactually mean.",
                color = Color.White,
                fontSize = 28.sp,
                fontWeight = FontWeight.Medium,
                lineHeight = 36.sp
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = "Anonymous conversations with strangers.\nOne prompt. 20 minutes. No traces.",
                color = Color.White.copy(alpha = 0.5f),
                fontSize = 16.sp,
                lineHeight = 24.sp
            )
        }
        RaazButton(text = "Let's Begin", onClick = onNext)
    }
}

@Composable
private fun CategoryPage(
    selectedCategory: String,
    onSelect: (String) -> Unit,
    onNext: () -> Unit
) {
    Column(
        modifier = Modifier.fillMaxSize(),
        verticalArrangement = Arrangement.SpaceBetween
    ) {
        Column {
            Text(
                text = "What do you want\nto talk about?",
                color = Color.White,
                fontSize = 28.sp,
                fontWeight = FontWeight.Bold,
                lineHeight = 36.sp
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "Pick one. You can change it later.",
                color = Color.White.copy(alpha = 0.4f),
                fontSize = 14.sp
            )
            Spacer(modifier = Modifier.height(32.dp))
            CATEGORIES.forEach { category ->
                SelectableChip(
                    text = category,
                    selected = selectedCategory == category,
                    onClick = { onSelect(category) }
                )
                Spacer(modifier = Modifier.height(12.dp))
            }
        }
        RaazButton(
            text = "Next",
            onClick = onNext,
            enabled = selectedCategory.isNotEmpty()
        )
    }
}

@Composable
private fun DemographicsPage(
    selectedAgeBracket: String,
    selectedCity: String,
    onSelectAge: (String) -> Unit,
    onSelectCity: (String) -> Unit,
    onComplete: () -> Unit
) {
    Column(
        modifier = Modifier.fillMaxSize(),
        verticalArrangement = Arrangement.SpaceBetween
    ) {
        Column {
            Text(
                text = "A bit about you",
                color = Color.White,
                fontSize = 28.sp,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "Used only for matching. Never shown publicly.",
                color = Color.White.copy(alpha = 0.4f),
                fontSize = 14.sp
            )
            Spacer(modifier = Modifier.height(32.dp))
            Text(
                text = "AGE",
                color = Color.White.copy(alpha = 0.5f),
                fontSize = 11.sp,
                fontWeight = FontWeight.Medium,
                letterSpacing = 1.5.sp
            )
            Spacer(modifier = Modifier.height(8.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                AGE_BRACKETS.forEach { bracket ->
                    SelectableChip(
                        text = bracket,
                        selected = selectedAgeBracket == bracket,
                        onClick = { onSelectAge(bracket) }
                    )
                }
            }
            Spacer(modifier = Modifier.height(24.dp))
            Text(
                text = "CITY",
                color = Color.White.copy(alpha = 0.5f),
                fontSize = 11.sp,
                fontWeight = FontWeight.Medium,
                letterSpacing = 1.5.sp
            )
            Spacer(modifier = Modifier.height(8.dp))
            CITIES.forEach { city ->
                SelectableChip(
                    text = city,
                    selected = selectedCity == city,
                    onClick = { onSelectCity(city) }
                )
                Spacer(modifier = Modifier.height(12.dp))
            }
        }
        RaazButton(
            text = "Start Raaz",
            onClick = onComplete,
            enabled = selectedAgeBracket.isNotEmpty() && selectedCity.isNotEmpty()
        )
    }
}

@Composable
fun SelectableChip(text: String, selected: Boolean, onClick: () -> Unit) {
    val bg = if (selected) RaazAccent else Color.Transparent
    val textColor = if (selected) Color.Black else Color.White
    val borderColor = if (selected) RaazAccent else Color.White.copy(alpha = 0.2f)

    Box(
        modifier = Modifier
            .background(bg, shape = RoundedCornerShape(8.dp))
            .border(1.dp, borderColor, shape = RoundedCornerShape(8.dp))
            .clickable(onClick = onClick)
            .padding(horizontal = 20.dp, vertical = 12.dp)
    ) {
        Text(
            text = text,
            color = textColor,
            fontSize = 15.sp,
            fontWeight = FontWeight.Medium
        )
    }
}

@Composable
fun RaazButton(
    text: String,
    onClick: () -> Unit,
    enabled: Boolean = true,
    modifier: Modifier = Modifier
) {
    Button(
        onClick = onClick,
        enabled = enabled,
        modifier = modifier
            .fillMaxWidth()
            .height(52.dp),
        shape = RoundedCornerShape(10.dp),
        colors = ButtonDefaults.buttonColors(
            containerColor = RaazAccent,
            contentColor = Color.Black,
            disabledContainerColor = Color.White.copy(alpha = 0.08f),
            disabledContentColor = Color.White.copy(alpha = 0.25f)
        )
    ) {
        Text(text = text, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
    }
}
