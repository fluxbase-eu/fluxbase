package io.github.nimbleflux.fluxbase.graphql

import io.github.nimbleflux.fluxbase.FluxbaseClient
import io.github.nimbleflux.fluxbase.FluxbaseClientOptions
import io.github.nimbleflux.fluxbase.core.test.RecordingHttp
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.JsonObject
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class FluxbaseGraphQLTest {

    private fun client(recording: RecordingHttp): FluxbaseClient =
        FluxbaseClient.create("http://localhost:8080", "anon-key", FluxbaseClientOptions(autoRefresh = false), recording)

    @Serializable
    private data class Range(val from: String, val to: String)

    @Test
    fun `query posts query and variables to graphql endpoint`() = runTest {
        val recording = RecordingHttp(mockResponseBody = """{"data":{"trips":[]}}""")
        val c = client(recording)

        c.graphql.query<List<Map<String, String>>>("query { trips { id } }", mapOf("limit" to 5))

        assertEquals("POST", recording.lastMethod)
        assertEquals("/api/v1/graphql", recording.lastPath)
        val body = recording.lastBody as JsonObject
        assertEquals("query { trips { id } }", body["query"].toString().trim('"'))
    }

    @Test
    fun `variables encode non-primitives as real JSON`() = runTest {
        val recording = RecordingHttp(mockResponseBody = """{"data":{}}""")
        val c = client(recording)

        c.graphql.query<Map<String, String>>(
            "query Search(${'$'}tags: [String!]!) { search(tags: ${'$'}tags) { id } }",
            mapOf(
                "tags" to listOf("restaurant", "cafe"),
                "range" to Range(from = "2026-01-01", to = "2026-02-01"),
                "cursor" to null,
            ),
        )

        val variables = (recording.lastBody as JsonObject)["variables"] as JsonObject
        // Lists encode as JSON arrays (previously the string "[restaurant, cafe]")
        assertEquals("""["restaurant","cafe"]""", variables["tags"].toString())
        // @Serializable dataclasses encode via their generated serializer
        assertEquals("""{"from":"2026-01-01","to":"2026-02-01"}""", variables["range"].toString())
        assertTrue(variables["cursor"] is JsonNull)
    }
}
