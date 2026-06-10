package com.raaz.app.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

private val RaazDarkColorScheme = darkColorScheme(
    primary = RaazAccent,
    onPrimary = Color.Black,
    background = RaazBackground,
    onBackground = Color.White,
    surface = RaazSurface,
    onSurface = Color.White,
    surfaceVariant = Color(0xFF1A1A1A),
    onSurfaceVariant = Color(0x99FFFFFF),
    outline = Color(0x1FFFFFFF),
    error = RaazError
)

@Composable
fun RaazTheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colorScheme = RaazDarkColorScheme,
        typography = RaazTypography,
        content = content
    )
}
