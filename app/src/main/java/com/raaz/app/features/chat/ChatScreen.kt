package com.raaz.app.features.chat

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.systemBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
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
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.text.style.TextAlign
import androidx.lifecycle.viewmodel.compose.viewModel
import com.raaz.app.BuildConfig
import com.raaz.app.data.models.ChatMessage
import com.raaz.app.data.websocket.ConnectionState
import com.raaz.app.ui.theme.RaazAccent
import com.raaz.app.ui.theme.RaazBackground
import com.raaz.app.ui.theme.RaazSurface

@Composable
fun ChatScreen(
    prompt: String,
    onBack: () -> Unit,
    viewModel: ChatViewModel = run {
        val context = LocalContext.current
        val application = context.applicationContext as android.app.Application
        val factory = remember { ChatViewModelFactory(application, BuildConfig.WS_BASE_URL, prompt) }
        viewModel(factory = factory)
    }
) {
    val timerText by viewModel.timerText.collectAsState()
    val deletionCountdown by viewModel.deletionCountdown.collectAsState()
    val inputText by viewModel.inputText.collectAsState()
    val sessionAlias by viewModel.sessionAlias.collectAsState()
    val messages by viewModel.messages.collectAsState()
    val isTyping by viewModel.isTyping.collectAsState()
    val partnerTyping by viewModel.partnerTyping.collectAsState()
    val extensionRequest by viewModel.extensionRequest.collectAsState()
    val partnerHandleExchange by viewModel.partnerHandleExchange.collectAsState()
    val userApprovedHandle by viewModel.userApprovedHandle.collectAsState()
    val partnerHandle by viewModel.partnerHandle.collectAsState()
    val connectionState by viewModel.connectionState.collectAsState()
    val pendingVaultMessage by viewModel.pendingVaultMessage.collectAsState()
    val moderationAlert by viewModel.moderationAlert.collectAsState()
    val crisisTriggered by viewModel.crisisTriggered.collectAsState()
    val userHandle = viewModel.getVisibleUserHandle()

    LaunchedEffect(Unit) {
        viewModel.vaultSaveSuccess.collect {
            // Vault save confirmed — could show snackbar here in future
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(RaazBackground)
            .systemBarsPadding()
    ) {
        if (connectionState == ConnectionState.RECONNECTING) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color(0xFFFFA500).copy(alpha = 0.12f))
                    .padding(horizontal = 16.dp, vertical = 6.dp),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "Reconnecting...",
                    color = Color(0xFFFFA500),
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Medium
                )
            }
        }

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
                        text = if (partnerTyping) "typing..." else prompt,
                        color = if (partnerTyping) RaazAccent else Color.White.copy(alpha = 0.35f),
                        fontSize = 11.sp,
                        maxLines = 1
                    )
                    if (deletionCountdown.isNotEmpty()) {
                        Text(
                            text = deletionCountdown,
                            color = Color.White.copy(alpha = 0.22f),
                            fontSize = 10.sp
                        )
                    }
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
            contentPadding = PaddingValues(vertical = 32.dp),
            reverseLayout = false
        ) {
            if (messages.isEmpty()) {
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
            } else {
                items(messages, key = { it.messageId }) { message ->
                    MessageBubble(
                        message = message,
                        onLongClick = if (message.isOwnMessage) {
                            { viewModel.requestVaultSave(message) }
                        } else null
                    )
                }
            }
            if (isTyping) {
                item {
                    Box(
                        modifier = Modifier.fillMaxWidth(),
                        contentAlignment = Alignment.CenterStart
                    ) {
                        Text(
                            text = "typing...",
                            color = Color.White.copy(alpha = 0.4f),
                            fontSize = 12.sp
                        )
                    }
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
                onClick = { viewModel.requestExtension("partner-id") },
                enabled = extensionRequest == null,
                modifier = Modifier.weight(1f),
                shape = RoundedCornerShape(8.dp),
                border = BorderStroke(1.dp, Color.White.copy(alpha = 0.08f)),
                colors = ButtonDefaults.outlinedButtonColors(
                    disabledContentColor = Color.White.copy(alpha = 0.25f),
                    contentColor = Color.White
                )
            ) {
                Text(text = if (extensionRequest != null) "Pending..." else "+10 min", fontSize = 12.sp)
            }
            OutlinedButton(
                onClick = { viewModel.initiateHandleExchange("partner-id") },
                enabled = !userApprovedHandle,
                modifier = Modifier.weight(1f),
                shape = RoundedCornerShape(8.dp),
                border = BorderStroke(1.dp, Color.White.copy(alpha = 0.08f)),
                colors = ButtonDefaults.outlinedButtonColors(
                    disabledContentColor = Color.White.copy(alpha = 0.25f),
                    contentColor = Color.White
                )
            ) {
                Text(
                    text = if (userApprovedHandle) "Waiting..." else "Exchange Handle",
                    fontSize = 12.sp
                )
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

    if (pendingVaultMessage != null) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black.copy(alpha = 0.5f)),
            contentAlignment = Alignment.Center
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth(0.8f)
                    .background(RaazSurface, RoundedCornerShape(12.dp))
                    .padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Text(
                    text = "Save to Vault?",
                    color = Color.White,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.SemiBold
                )
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = "This message will be stored locally on your device. You can delete it anytime from your Vault. Your consent is required under DPDP Act 2023.",
                    color = Color.White.copy(alpha = 0.55f),
                    fontSize = 12.sp,
                    lineHeight = 18.sp
                )
                Spacer(modifier = Modifier.height(20.dp))
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    OutlinedButton(
                        onClick = { viewModel.cancelVaultSave() },
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(8.dp),
                        border = BorderStroke(1.dp, Color.White.copy(alpha = 0.3f)),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = Color.White)
                    ) {
                        Text(text = "Cancel", fontSize = 12.sp)
                    }
                    OutlinedButton(
                        onClick = { viewModel.confirmVaultSave() },
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(8.dp),
                        border = BorderStroke(1.dp, RaazAccent),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = RaazAccent)
                    ) {
                        Text(text = "Save to Vault", fontSize = 12.sp)
                    }
                }
            }
        }
    }

    if (extensionRequest != null) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black.copy(alpha = 0.5f)),
            contentAlignment = Alignment.Center
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth(0.8f)
                    .background(RaazSurface, RoundedCornerShape(12.dp))
                    .padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Text(
                    text = "${extensionRequest!!.requesterAlias} wants to extend chat by 10 minutes",
                    color = Color.White,
                    fontSize = 14.sp,
                    lineHeight = 20.sp
                )
                Spacer(modifier = Modifier.padding(16.dp))
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    OutlinedButton(
                        onClick = { viewModel.respondToExtension(false) },
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(8.dp),
                        border = BorderStroke(1.dp, Color.White.copy(alpha = 0.3f)),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = Color.White)
                    ) {
                        Text(text = "Decline", fontSize = 12.sp)
                    }
                    OutlinedButton(
                        onClick = { viewModel.respondToExtension(true) },
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(8.dp),
                        border = BorderStroke(1.dp, RaazAccent),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = RaazAccent)
                    ) {
                        Text(text = "Approve", fontSize = 12.sp)
                    }
                }
            }
        }
    }

    if (partnerHandleExchange != null) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black.copy(alpha = 0.5f)),
            contentAlignment = Alignment.Center
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth(0.8f)
                    .background(RaazSurface, RoundedCornerShape(12.dp))
                    .padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Text(
                    text = "${partnerHandleExchange!!.userAlias} wants to exchange contact handles",
                    color = Color.White,
                    fontSize = 14.sp,
                    lineHeight = 20.sp
                )
                Spacer(modifier = Modifier.padding(16.dp))
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    OutlinedButton(
                        onClick = { viewModel.respondToHandleExchange(false) },
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(8.dp),
                        border = BorderStroke(1.dp, Color.White.copy(alpha = 0.3f)),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = Color.White)
                    ) {
                        Text(text = "Decline", fontSize = 12.sp)
                    }
                    OutlinedButton(
                        onClick = { viewModel.respondToHandleExchange(true) },
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(8.dp),
                        border = BorderStroke(1.dp, RaazAccent),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = RaazAccent)
                    ) {
                        Text(text = "Approve", fontSize = 12.sp)
                    }
                }
            }
        }
    }

    if (userHandle != null && partnerHandle != null) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(RaazAccent.copy(alpha = 0.1f))
                .padding(12.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(text = "Handles Exchanged", color = RaazAccent, fontSize = 11.sp, fontWeight = FontWeight.SemiBold)
            Row(
                horizontalArrangement = Arrangement.spacedBy(16.dp),
                modifier = Modifier.padding(top = 8.dp)
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(text = "You", color = Color.White.copy(alpha = 0.6f), fontSize = 10.sp)
                    Text(text = userHandle!!, color = Color.White, fontSize = 12.sp, fontWeight = FontWeight.SemiBold)
                }
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(text = "Partner", color = Color.White.copy(alpha = 0.6f), fontSize = 10.sp)
                    Text(text = partnerHandle!!, color = Color.White, fontSize = 12.sp, fontWeight = FontWeight.SemiBold)
                }
            }
        }
    }

    if (moderationAlert != null) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black.copy(alpha = 0.55f)),
            contentAlignment = Alignment.Center
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth(0.85f)
                    .background(RaazSurface, RoundedCornerShape(12.dp))
                    .padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Text(
                    text = if (moderationAlert!!.action == "disconnect") "You've been disconnected" else "Message blocked",
                    color = Color.White,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.SemiBold
                )
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = moderationAlert!!.reason,
                    color = Color.White.copy(alpha = 0.6f),
                    fontSize = 13.sp,
                    textAlign = TextAlign.Center,
                    lineHeight = 20.sp
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "Strike ${moderationAlert!!.strikeNum} of 2",
                    color = Color(0xFFFFA500).copy(alpha = 0.85f),
                    fontSize = 11.sp
                )
                Spacer(modifier = Modifier.height(20.dp))
                OutlinedButton(
                    onClick = { viewModel.dismissModerationAlert() },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(8.dp),
                    border = BorderStroke(1.dp, Color.White.copy(alpha = 0.3f)),
                    colors = ButtonDefaults.outlinedButtonColors(contentColor = Color.White)
                ) {
                    Text(text = "OK", fontSize = 13.sp)
                }
            }
        }
    }

    // Crisis overlay renders last so it sits on top of all other overlays.
    if (crisisTriggered != null) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color(0xFF050505)),
            contentAlignment = Alignment.Center
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 36.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                Text(
                    text = "You matter.",
                    color = Color.White,
                    fontSize = 28.sp,
                    fontWeight = FontWeight.Bold,
                    textAlign = TextAlign.Center
                )
                Text(
                    text = crisisTriggered!!.message,
                    color = Color.White.copy(alpha = 0.65f),
                    fontSize = 14.sp,
                    textAlign = TextAlign.Center,
                    lineHeight = 22.sp
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "Helplines",
                    color = RaazAccent,
                    fontSize = 12.sp,
                    fontWeight = FontWeight.SemiBold
                )
                crisisTriggered!!.helplines.forEach { helpline ->
                    Text(
                        text = helpline,
                        color = Color.White.copy(alpha = 0.9f),
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Medium,
                        textAlign = TextAlign.Center
                    )
                }
                Spacer(modifier = Modifier.height(24.dp))
                OutlinedButton(
                    onClick = { viewModel.dismissCrisis() },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(8.dp),
                    border = BorderStroke(1.dp, RaazAccent),
                    colors = ButtonDefaults.outlinedButtonColors(contentColor = RaazAccent)
                ) {
                    Text(text = "I'm OK", fontSize = 14.sp, fontWeight = FontWeight.SemiBold)
                }
            }
        }
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
fun MessageBubble(
    message: ChatMessage,
    modifier: Modifier = Modifier,
    onLongClick: (() -> Unit)? = null
) {
    Column(
        modifier = modifier.fillMaxWidth(),
        horizontalAlignment = if (message.isOwnMessage) Alignment.End else Alignment.Start
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth(0.8f)
                .background(
                    if (message.isOwnMessage) RaazAccent.copy(alpha = 0.2f) else RaazSurface,
                    RoundedCornerShape(12.dp)
                )
                .then(
                    if (onLongClick != null) {
                        Modifier.combinedClickable(onClick = {}, onLongClick = onLongClick)
                    } else {
                        Modifier
                    }
                )
                .padding(12.dp)
        ) {
            Column {
                if (!message.isOwnMessage) {
                    Text(
                        text = message.senderAlias,
                        color = RaazAccent,
                        fontSize = 10.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                }
                Text(
                    text = message.text,
                    color = Color.White,
                    fontSize = 14.sp,
                    lineHeight = 20.sp
                )
            }
        }
    }
}
