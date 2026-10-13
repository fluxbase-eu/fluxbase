package io.github.nimbleflux.fluxbase.ai

import io.github.nimbleflux.fluxbase.FluxbaseResponse
import io.github.nimbleflux.fluxbase.core.FluxbaseHttpClient
import io.github.nimbleflux.fluxbase.fluxbaseResponse
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

/**
 * AI module — chatbots, conversations, and usage. Mirrors the user AI
 * surface of the TS SDK (`sdk/src/ai.ts`, `FluxbaseAI`). The WebSocket chat
 * stream is not covered here; use the TS/React SDKs or the `/ai/ws` protocol
 * directly for streaming chat.
 */
class FluxbaseAI(private val http: FluxbaseHttpClient) {

    private fun encode(s: String): String = java.net.URLEncoder.encode(s, "UTF-8")

    /** GET `/api/v1/ai/chatbots` — list public/enabled chatbots. */
    suspend fun listChatbots(namespace: String? = null): FluxbaseResponse<List<AIChatbotSummary>> =
        fluxbaseResponse {
            val query = namespace?.let { "?namespace=${encode(it)}" } ?: ""
            http.get("/api/v1/ai/chatbots$query")
        }

    /** GET `/api/v1/ai/chatbots/:id` — get a chatbot by ID. */
    suspend fun getChatbot(id: String): FluxbaseResponse<AIChatbotSummary> = fluxbaseResponse {
        http.get("/api/v1/ai/chatbots/$id")
    }

    /** GET `/api/v1/ai/chatbots/by-name/:name?namespace=…` — resolve a chatbot by name. */
    suspend fun lookupChatbot(name: String, namespace: String = "default"): FluxbaseResponse<AIChatbotSummary> =
        fluxbaseResponse {
            http.get("/api/v1/ai/chatbots/by-name/${encode(name)}?namespace=${encode(namespace)}")
        }

    /** GET `/api/v1/ai/conversations` — list the current user's conversations. */
    suspend fun listConversations(
        chatbotId: String? = null,
        limit: Int? = null,
        offset: Int? = null,
    ): FluxbaseResponse<AIConversationListResult> = fluxbaseResponse {
        val params = mutableListOf<String>()
        chatbotId?.let { params.add("chatbot_id=${encode(it)}") }
        limit?.let { params.add("limit=$it") }
        offset?.let { params.add("offset=$it") }
        val query = if (params.isEmpty()) "" else "?" + params.joinToString("&")
        http.get("/api/v1/ai/conversations$query")
    }

    /** GET `/api/v1/ai/conversations/:id` — conversation with full message history. */
    suspend fun getConversation(id: String): FluxbaseResponse<AIConversationDetail> =
        fluxbaseResponse {
            http.get("/api/v1/ai/conversations/$id")
        }

    /** PATCH `/api/v1/ai/conversations/:id` — update a conversation (e.g. rename). */
    suspend fun updateConversation(
        id: String,
        request: UpdateConversationRequest,
    ): FluxbaseResponse<AIConversationDetail> = fluxbaseResponse {
        val body = buildJsonObject { request.title?.let { put("title", it) } }
        http.patch("/api/v1/ai/conversations/$id", body)
    }

    /** DELETE `/api/v1/ai/conversations/:id` — delete a conversation. */
    suspend fun deleteConversation(id: String): FluxbaseResponse<Unit> = fluxbaseResponse {
        http.delete("/api/v1/ai/conversations/$id")
        Unit
    }

    /** GET `/api/v1/ai/usage/:chatbotId` — per-user daily usage snapshot. */
    suspend fun getUsage(chatbotId: String): FluxbaseResponse<AIUsageSnapshot> = fluxbaseResponse {
        http.get("/api/v1/ai/usage/$chatbotId")
    }
}
