package com.raaz.app.features.chat

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.systemBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.BorderStroke
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.material3.TextFieldDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.raaz.app.ui.theme.RaazAccent
import com.raaz.app.ui.theme.RaazBackground
import com.raaz.app.ui.theme.RaazSurface

@Composable
fun ChatScreen(
    prompt: String,
    onBack: () -> Unit,
    viewModel: ChatViewModel = viewModel()
) {
    val timerText by viewModel.timerText.collectAsState()
    val inputText by viewModel.inputText.collectAsState()
    val sessionAlias by viewModel.sessionAlias.collectAsState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(RaazBackground)
            .systemBarsPadding()
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(end = 16.dp, top = 4.dp, bottom = 4.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                IconButton(onClick = onBack) {
                    Icon(
                        imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                        contentDescription = "Back",
                        tint = Color.White
                    )
                }
                Column {
                    Text(
                        text = sessionAlias,
                        color = Color.White,
                        fontSize = 15.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                    Text(
                        text = prompt,
                        color = Color.White.copy(alpha = 0.35f),
                        fontSize = 11.sp,
                        maxLines = 1
                    )
                }
            }
            TimerComponent(timerText = timerText)
        }

        HorizontalDivider(color = Color.White.copy(alpha = 0.06f))

        LazyColumn(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth()
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
            contentPadding = PaddingValues(vertical = 32.dp)
        ) {
            item {
                Box(
                    modifier = Modifier.fillMaxWidth(),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "Conversation begins. Be honest.",
                        color = Color.White.copy(alpha = 0.2f),
                        fontSize = 12.sp
                    )
                }
            }
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 8.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            OutlinedButton(
                onClick = {},
                enabled = false,
                modifier = Modifier.weight(1f),
                shape = RoundedCornerShape(8.dp),
                border = BorderStroke(1.dp, Color.White.copy(alpha = 0.08f)),
                colors = ButtonDefaults.outlinedButtonColors(
                    disabledContentColor = Color.White.copy(alpha = 0.25f)
                )
            ) {
                Text(text = "+10 min", fontSize = 12.sp)
            }
            OutlinedButton(
                onClick = {},
                enabled = false,
                modifier = Modifier.weight(1f),
                shape = RoundedCornerShape(8.dp),
                border = BorderStroke(1.dp, Color.White.copy(alpha = 0.08f)),
                colors = ButtonDefaults.outlinedButtonColors(
                    disabledContentColor = Color.White.copy(alpha = 0.25f)
                )
            ) {
                Text(text = "Exchange Handle", fontSize = 12.sp)
            }
        }

        HorizontalDivider(color = Color.White.copy(alpha = 0.06f))

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 12.dp, vertical = 8.dp)
                .imePadding(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            TextField(
                value = inputText,
                onValueChange = viewModel::onInputChange,
                modifier = Modifier.weight(1f),
                placeholder = {
                    Text(
                        text = "Say something real...",
                        color = Color.White.copy(alpha = 0.2f)
                    )
                },
                colors = TextFieldDefaults.colors(
                    focusedContainerColor = RaazSurface,
                    unfocusedContainerColor = RaazSurface,
                    focusedTextColor = Color.White,
                    unfocusedTextColor = Color.White,
                    focusedIndicatorColor = Color.Transparent,
                    unfocusedIndicatorColor = Color.Transparent,
                    cursorColor = RaazAccent
                ),
                shape = RoundedCornerShape(10.dp),
                maxLines = 4
            )
            IconButton(
                onClick = viewModel::sendMessage,
                modifier = Modifier
                    .background(RaazAccent, RoundedCornerShape(10.dp))
                    .size(48.dp)
            ) {
                Text(
                    text = "↑",
                    color = Color.Black,
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold
                )
            }
        }
    }
}
