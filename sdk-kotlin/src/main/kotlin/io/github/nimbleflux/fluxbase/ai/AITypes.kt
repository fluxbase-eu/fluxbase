package io.github.nimbleflux.fluxbase.ai

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class AIChatbotSummary(
    val id: String,
    val name: String,
    val namespace: String = "default",
    val description: String? = null,
    val model: String? = null,
    val enabled: Boolean = true,
    @SerialName("is_public") val isPublic: Boolean = true,
    val source: String = "api",
    @SerialName("updated_at") val updatedAt: String? = null,
)

/** One counter half of [AIUsageSnapshot]. */
@Serializable
data class AIQuota(
    val used: Long = 0,
    val limit: Long = 0,
)

/** Per-user daily usage snapshot for a chatbot (GET /api/v1/ai/usage/:chatbotId). */
@Serializable
data class AIUsageSnapshot(
    val requests: AIQuota? = null,
    val tokens: AIQuota? = null,
    @SerialName("resets_at") val resetsAt: String? = null,
)

/** User-view conversation summary (GET /api/v1/ai/conversations). */
@Serializable
data class AIConversationSummary(
    val id: String,
    /** Chatbot *name* (not ID) in the user API. */
    val chatbot: String? = null,
    val namespace: String? = null,
    val title: String? = null,
    val preview: String? = null,
    @SerialName("message_count") val messageCount: Int = 0,
    @SerialName("created_at") val createdAt: String? = null,
    @SerialName("updated_at") val updatedAt: String? = null,
)

@Serializable
data class AIConversationListResult(
    val conversations: List<AIConversationSummary> = emptyList(),
    val total: Int = 0,
    @SerialName("has_more") val hasMore: Boolean = false,
)

@Serializable
data class AIConversationMessage(
    val id: String? = null,
    @SerialName("conversation_id") val conversationId: String? = null,
    val role: String,
    val content: String,
    @SerialName("created_at") val createdAt: String? = null,
)

@Serializable
data class AIConversationDetail(
    val id: String,
    @SerialName("chatbot_id") val chatbotId: String? = null,
    val title: String? = null,
    val status: String? = null,
    val messages: List<AIConversationMessage> = emptyList(),
    @SerialName("created_at") val createdAt: String? = null,
    @SerialName("updated_at") val updatedAt: String? = null,
)

@Serializable
data class UpdateConversationRequest(
    val title: String? = null,
)
