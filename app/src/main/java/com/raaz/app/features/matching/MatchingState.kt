package com.raaz.app.features.matching

sealed class MatchingState {
    object Idle : MatchingState()

    data class InQueue(
        val elapsedSeconds: Long,
        val remainingSeconds: Long = 900 - elapsedSeconds
    ) : MatchingState()

    data class TimeoutFallback(
        val elapsedSeconds: Long,
        val reason: String = "National pool activated (city constraint dropped)"
    ) : MatchingState()

    data class Matched(
        val matchId: String,
        val partnerAlias: String
    ) : MatchingState()

    data class Error(val message: String) : MatchingState()
}
