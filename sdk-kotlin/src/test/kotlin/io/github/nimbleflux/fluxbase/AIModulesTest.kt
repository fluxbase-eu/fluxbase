package io.github.nimbleflux.fluxbase

import io.github.nimbleflux.fluxbase.ai.FluxbaseAI
import io.github.nimbleflux.fluxbase.ai.UpdateConversationRequest
import io.github.nimbleflux.fluxbase.core.FluxbaseHttpClient
import io.github.nimbleflux.fluxbase.core.test.RecordingHttp
import io.github.nimbleflux.fluxbase.knowledgebase.CreateKnowledgeBaseRequest
import io.github.nimbleflux.fluxbase.knowledgebase.FluxbaseKnowledgeBase
import io.github.nimbleflux.fluxbase.knowledgebase.GrantKBPermissionRequest
import io.github.nimbleflux.fluxbase.knowledgebase.KBSearchRequest
import io.github.nimbleflux.fluxbase.vector.VectorSearchMetric
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertTrue

/**
 * Wire-shape tests for the AI and KnowledgeBase modules — same approach as
 * the auth tests: a [RecordingHttp] fake asserts the exact paths, methods,
 * and bodies the modules produce.
 */
class AIModulesTest {

    private fun http(recording: RecordingHttp) = FluxbaseHttpClient("http://localhost:8080", recording)

    /** The transport normally JSON-encodes request bodies; mirror that here. */
    private fun recordedBody(recording: RecordingHttp): JsonObject =
        io.github.nimbleflux.fluxbase.core.KtorHttpTransport("http://localhost:8080")
            .encodeToJsonElement(recording.lastBody)
            .jsonObject

    // ------------------------------------------------------------------
    // FluxbaseAI
    // ------------------------------------------------------------------

    @Test
    fun `listChatbots GETs the chatbots route with optional namespace`() = runTest {
        val recording = RecordingHttp(mockResponseBody = """[{"id":"bot-1","name":"assistant"}]""")
        val ai = FluxbaseAI(http(recording))

        val (data, err) = ai.listChatbots()
        assertEquals(null, err)
        assertEquals("GET", recording.lastMethod)
        assertEquals("/api/v1/ai/chatbots", recording.lastPath)
        assertEquals("bot-1", data!!.first().id)

        ai.listChatbots(namespace = "support")
        assertEquals("/api/v1/ai/chatbots?namespace=support", recording.lastPath)
    }

    @Test
    fun `lookupChatbot hits the by-name route`() = runTest {
        val recording = RecordingHttp(mockResponseBody = """{"id":"bot-1","name":"assistant"}""")
        val ai = FluxbaseAI(http(recording))

        ai.lookupChatbot("assistant", namespace = "support")

        assertEquals("/api/v1/ai/chatbots/by-name/assistant?namespace=support", recording.lastPath)
    }

    @Test
    fun `listConversations builds query params`() = runTest {
        val recording = RecordingHttp(
            mockResponseBody = """{"conversations":[{"id":"c-1"}],"total":1,"has_more":false}""",
        )
        val ai = FluxbaseAI(http(recording))

        val (data, err) = ai.listConversations(chatbotId = "bot-1", limit = 10, offset = 5)
        assertEquals(null, err)
        assertEquals(1, data!!.total)
        assertEquals(
            "/api/v1/ai/conversations?chatbot_id=bot-1&limit=10&offset=5",
            recording.lastPath,
        )
    }

    @Test
    fun `updateConversation PATCHes the title`() = runTest {
        val recording = RecordingHttp(mockResponseBody = """{"id":"c-1","title":"renamed"}""")
        val ai = FluxbaseAI(http(recording))

        ai.updateConversation("c-1", UpdateConversationRequest(title = "renamed"))

        assertEquals("PATCH", recording.lastMethod)
        assertEquals("/api/v1/ai/conversations/c-1", recording.lastPath)
        val body = recordedBody(recording)
        assertEquals("renamed", body["title"]!!.jsonPrimitive.content)
    }

    @Test
    fun `deleteConversation DELETEs the conversation`() = runTest {
        val recording = RecordingHttp()
        val ai = FluxbaseAI(http(recording))

        val response = ai.deleteConversation("c-1")
        assertEquals(null, response.error)
        assertEquals(Unit, response.data)
        assertEquals("DELETE", recording.lastMethod)
        assertEquals("/api/v1/ai/conversations/c-1", recording.lastPath)
    }

    @Test
    fun `getUsage GETs the usage route`() = runTest {
        val recording = RecordingHttp(
            mockResponseBody = """{"requests":{"used":3,"limit":500},"tokens":{"used":100,"limit":100000},"resets_at":"2026-10-13T00:00:00Z"}""",
        )
        val ai = FluxbaseAI(http(recording))

        val (data, err) = ai.getUsage("bot-1")
        assertEquals(null, err)
        assertEquals("/api/v1/ai/usage/bot-1", recording.lastPath)
        assertEquals(3L, data!!.requests!!.used)
        assertEquals(500L, data.requests!!.limit)
    }

    // ------------------------------------------------------------------
    // FluxbaseKnowledgeBase
    // ------------------------------------------------------------------

    @Test
    fun `create POSTs the KB create route with new fields`() = runTest {
        val recording = RecordingHttp(mockResponseBody = """{"id":"kb-1","name":"docs","rerank_enabled":false}""")
        val kb = FluxbaseKnowledgeBase(http(recording))

        val (data, err) = kb.create(
            CreateKnowledgeBaseRequest(
                name = "docs",
                entityExtractionStrategy = "hybrid",
                rerankEnabled = false,
            ),
        )
        assertEquals(null, err)
        assertEquals("kb-1", data!!.id)
        assertEquals("POST", recording.lastMethod)
        assertEquals("/api/v1/ai/knowledge-bases", recording.lastPath)

        val body = recordedBody(recording)
        assertEquals("docs", body["name"]!!.jsonPrimitive.content)
        assertEquals("hybrid", body["entity_extraction_strategy"]!!.jsonPrimitive.content)
        assertEquals("false", body["rerank_enabled"]!!.jsonPrimitive.content)
    }

    @Test
    fun `update PATCHes the user KB route`() = runTest {
        val recording = RecordingHttp(mockResponseBody = """{"id":"kb-1"}""")
        val kb = FluxbaseKnowledgeBase(http(recording))

        kb.update("kb-1", io.github.nimbleflux.fluxbase.knowledgebase.UpdateKnowledgeBaseRequest(rerankEnabled = true))

        assertEquals("PATCH", recording.lastMethod)
        assertEquals("/api/v1/ai/knowledge-bases/kb-1", recording.lastPath)
    }

    @Test
    fun `delete DELETEs the user KB route`() = runTest {
        val recording = RecordingHttp()
        val kb = FluxbaseKnowledgeBase(http(recording))

        val response = kb.delete("kb-1")
        assertEquals(null, response.error)
        assertEquals(Unit, response.data)
        assertEquals("DELETE", recording.lastMethod)
    }

    @Test
    fun `share grants and revokePermission deletes`() = runTest {
        val recording = RecordingHttp(
            mockResponseBody = """{"id":"perm-1","user_id":"u-2","permission":"viewer"}""",
        )
        val kb = FluxbaseKnowledgeBase(http(recording))

        val (grant, err) = kb.share("kb-1", GrantKBPermissionRequest(userId = "u-2", permission = "viewer"))
        assertEquals(null, err)
        assertEquals("u-2", grant!!.userId)
        assertEquals("POST", recording.lastMethod)
        assertEquals("/api/v1/ai/knowledge-bases/kb-1/share", recording.lastPath)

        kb.revokePermission("kb-1", "u-2")
        assertEquals("DELETE", recording.lastMethod)
        assertEquals("/api/v1/ai/knowledge-bases/kb-1/permissions/u-2", recording.lastPath)
    }

    @Test
    fun `search POSTs the aligned request shape`() = runTest {
        val recording = RecordingHttp(
            mockResponseBody = """[{"chunk_id":"ch-1","content":"hello","similarity":0.9}]""",
        )
        val kb = FluxbaseKnowledgeBase(http(recording))

        val (data, err) = kb.search(
            "kb-1",
            KBSearchRequest(query = "reset", limit = 5, mode = "hybrid", semanticWeight = 0.7, threshold = 0.5, rerank = true),
        )
        assertEquals(null, err)
        assertEquals("ch-1", data!!.first().chunkId)

        assertEquals("POST", recording.lastMethod)
        assertEquals("/api/v1/ai/knowledge-bases/kb-1/search", recording.lastPath)
        val body = recordedBody(recording)
        assertEquals("reset", body["query"]!!.jsonPrimitive.content)
        assertEquals(5, body["limit"]!!.jsonPrimitive.content.toInt())
        assertEquals("hybrid", body["mode"]!!.jsonPrimitive.content)
        assertEquals(0.7, body["semantic_weight"]!!.jsonPrimitive.content.toDouble())
        assertEquals("true", body["rerank"]!!.jsonPrimitive.content)
    }

    @Test
    fun `document upload builds a multipart body`() = runTest {
        val recording = RecordingHttp(mockResponseBody = """{"id":"doc-1","status":"pending"}""")
        val kb = FluxbaseKnowledgeBase(http(recording))

        val (data, err) = kb.uploadDocument(
            "kb-1",
            fileName = "guide.txt",
            content = "hello world".toByteArray(),
            mimeType = "text/plain",
            metadata = mapOf("tags" to "docs"),
        )
        assertEquals(null, err)
        assertEquals("doc-1", data!!.id)

        assertEquals("POST", recording.lastMethod)
        assertEquals("/api/v1/ai/knowledge-bases/kb-1/documents/upload", recording.lastPath)

        val headers = recording.lastHeaders
        val contentType = headers["Content-Type"]!!
        assertTrue(contentType.startsWith("multipart/form-data; boundary="), contentType)

        val bodyText = (recording.lastBody as ByteArray).decodeToString()
        assertTrue(bodyText.contains("Content-Disposition: form-data; name=\"tags\""), bodyText)
        assertTrue(bodyText.contains("Content-Disposition: form-data; name=\"file\"; filename=\"guide.txt\""), bodyText)
        assertTrue(bodyText.contains("Content-Type: text/plain"), bodyText)
        assertTrue(bodyText.contains("hello world"), bodyText)
    }

    @Test
    fun `graph and entity routes use the right paths`() = runTest {
        val recording = RecordingHttp(
            mockResponseBody = """{"entities":[],"relationships":[]}""",
        )
        val kb = FluxbaseKnowledgeBase(http(recording))

        kb.getKnowledgeGraph("kb-1")
        assertEquals("/api/v1/ai/knowledge-bases/kb-1/graph", recording.lastPath)

        kb.searchEntities("kb-1", "marie curie")
        // URLEncoder (used by the module) encodes spaces as '+'.
        assertEquals("/api/v1/ai/knowledge-bases/kb-1/entities/search?q=marie+curie", recording.lastPath)

        kb.listChatbots("kb-1")
        assertEquals("/api/v1/ai/knowledge-bases/kb-1/chatbots", recording.lastPath)
    }

    @Test
    fun `vector search gains filters and capabilities exists`() = runTest {
        val recording = RecordingHttp(
            mockResponseBody = """{"data":[],"distances":[],"model":"m"}""",
        )
        val vector = io.github.nimbleflux.fluxbase.vector.FluxbaseVector(http(recording))

        vector.search("documents", "embedding", query = "q", filters = mapOf("status" to "active"))
        val body = recording.lastBody as JsonObject
        val filters = body["filters"]!!.jsonObject
        assertEquals("active", filters["status"]!!.jsonPrimitive.content)
        assertEquals("cosine", body["metric"]!!.jsonPrimitive.content)
        assertEquals(VectorSearchMetric.COSINE.value, "cosine")

        recording.reset()
        recording.mockResponseBody = """{"enabled":true,"pgvector_version":"0.8.0"}"""
        val (caps, err) = vector.getCapabilities()
        assertEquals(null, err)
        assertEquals("/api/v1/capabilities/vector", recording.lastPath)
        assertEquals(true, caps!!.enabled)
    }
}
