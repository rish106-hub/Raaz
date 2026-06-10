package com.raaz.app.features.matching

import android.app.Application
import androidx.lifecycle.SavedStateHandle
import com.raaz.app.data.models.MatchRequest
import com.raaz.app.data.repository.AuthRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.Before
import org.junit.Test
import org.mockito.Mock
import org.mockito.MockitoAnnotations
import org.mockito.kotlin.whenever
import kotlin.test.assertEquals
import kotlin.test.assertTrue

@OptIn(ExperimentalCoroutinesApi::class)
class MatchingViewModelTest {

    @Mock
    private lateinit var mockApplication: Application

    @Mock
    private lateinit var mockAuthRepository: AuthRepository

    private lateinit var viewModel: MatchingViewModel
    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setUp() {
        MockitoAnnotations.openMocks(this)
        Dispatchers.setMain(testDispatcher)

        whenever(mockAuthRepository.getAnonymousAuth()).thenReturn(
            flowOf(AuthRepository.AuthState("test-session-123", "Raaz #1234"))
        )

        viewModel = MatchingViewModel(mockAuthRepository, mockApplication)
    }

    @Test
    fun testInitialStateIsIdle() {
        assertEquals(MatchingState.Idle, viewModel.matchingState.value)
    }

    @Test
    fun testStartMatchingEmitsInQueue() = runTest(testDispatcher) {
        viewModel.startMatching("test-prompt")
        advanceTimeBy(100)

        val state = viewModel.matchingState.value
        assertTrue(state is MatchingState.InQueue)
        assertTrue((state as MatchingState.InQueue).elapsedSeconds >= 0)
    }

    @Test
    fun testElapsedTimeIncrementsEachSecond() = runTest(testDispatcher) {
        viewModel.startMatching("test-prompt")

        advanceTimeBy(1000)
        val state1 = viewModel.matchingState.value
        if (state1 is MatchingState.InQueue) {
            assertTrue(state1.elapsedSeconds >= 1, "Expected elapsedSeconds >= 1, got ${state1.elapsedSeconds}")
        }

        advanceTimeBy(1000)
        val state2 = viewModel.matchingState.value
        if (state2 is MatchingState.InQueue) {
            assertTrue(state2.elapsedSeconds >= 2, "Expected elapsedSeconds >= 2, got ${state2.elapsedSeconds}")
        }
    }

    @Test
    fun testTransitionToTimeoutFallbackAfter900Seconds() = runTest(testDispatcher) {
        viewModel.startMatching("test-prompt")

        advanceTimeBy(900_000)

        val state = viewModel.matchingState.value
        assertTrue(state is MatchingState.TimeoutFallback, "Expected TimeoutFallback but got $state")
    }

    @Test
    fun testCancelMatchingReturnsToIdle() = runTest(testDispatcher) {
        viewModel.startMatching("test-prompt")
        advanceTimeBy(100)

        viewModel.cancelMatching()

        assertEquals(MatchingState.Idle, viewModel.matchingState.value)
    }

    @Test
    fun testCanRequestMatchValidation() {
        val validRequest = MatchRequest(
            promptId = "prompt",
            ageBucket = "20-25",
            city = "NYC",
            anonymousId = "id123"
        )

        val invalidRequest = MatchRequest(
            promptId = "",
            ageBucket = "20-25",
            city = "NYC",
            anonymousId = "id123"
        )

        assertTrue(viewModel.canRequestMatch(validRequest), "Valid request should pass")
        assertTrue(!viewModel.canRequestMatch(invalidRequest), "Invalid request should fail")
    }

    @Test
    fun testTimerCancelledOnViewModelCleared() = runTest(testDispatcher) {
        viewModel.startMatching("test-prompt")
        advanceTimeBy(100)

        viewModel.onCleared()

        advanceTimeBy(1000)
        val state = viewModel.matchingState.value
        assertTrue(
            state !is MatchingState.InQueue || (state as MatchingState.InQueue).elapsedSeconds < 2,
            "Timer should be cancelled, state should not update further"
        )
    }
}
