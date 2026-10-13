package io.github.nimbleflux.fluxbase

import kotlinx.coroutines.CoroutineStart
import kotlinx.coroutines.awaitCancellation
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue

class FluxbaseResponseTest {

    @Test
    fun `fluxbaseResponse wraps success`() = runTest {
        val result = fluxbaseResponse { "ok" }

        assertEquals("ok", result.data)
        assertNull(result.error)
    }

    @Test
    fun `fluxbaseResponse wraps exceptions as errors`() = runTest {
        val result = fluxbaseResponse<String> { throw IllegalStateException("boom") }

        assertNull(result.data)
        assertEquals("boom", result.error?.message)
    }

    @Test
    fun `fluxbaseResponse propagates caller cancellation`() = runTest {
        // CancellationException must escape the wrapper: swallowing it would let
        // the coroutine complete normally after its job was cancelled.
        val job = launch(start = CoroutineStart.UNDISPATCHED) {
            fluxbaseResponse<String> { awaitCancellation() }
        }

        assertTrue(job.isActive)
        job.cancel()
        assertTrue(job.isCancelled)
    }
}
