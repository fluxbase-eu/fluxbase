package io.github.nimbleflux.fluxbase.knowledgebase

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement

@Serializable
data class KnowledgeBaseSummary(
    val id: String,
    val name: String,
    val namespace: String = "default",
    val description: String? = null,
    val enabled: Boolean = true,
    @SerialName("document_count") val documentCount: Int = 0,
    @SerialName("total_chunks") val totalChunks: Int = 0,
    @SerialName("updated_at") val updatedAt: String? = null,
    val visibility: String? = null,
    @SerialName("user_permission") val userPermission: String? = null,
    @SerialName("owner_id") val ownerId: String? = null,
)

@Serializable
data class KnowledgeBase(
    val id: String,
    val name: String,
    val namespace: String = "default",
    val description: String? = null,
    @SerialName("embedding_model") val embeddingModel: String? = null,
    @SerialName("embedding_dimensions") val embeddingDimensions: Int = 1536,
    @SerialName("chunk_size") val chunkSize: Int = 512,
    @SerialName("chunk_overlap") val chunkOverlap: Int = 50,
    @SerialName("chunk_strategy") val chunkStrategy: String = "recursive",
    val enabled: Boolean = true,
    @SerialName("document_count") val documentCount: Int = 0,
    @SerialName("total_chunks") val totalChunks: Int = 0,
    val visibility: String? = null,
    @SerialName("entity_extraction_enabled") val entityExtractionEnabled: Boolean? = null,
    @SerialName("entity_extraction_strategy") val entityExtractionStrategy: String? = null,
    @SerialName("rerank_enabled") val rerankEnabled: Boolean? = null,
    @SerialName("created_at") val createdAt: String? = null,
    @SerialName("updated_at") val updatedAt: String? = null,
)

@Serializable
data class CreateKnowledgeBaseRequest(
    val name: String,
    val namespace: String? = null,
    val description: String? = null,
    val visibility: String? = null,
    @SerialName("embedding_model") val embeddingModel: String? = null,
    @SerialName("embedding_dimensions") val embeddingDimensions: Int? = null,
    @SerialName("chunk_size") val chunkSize: Int? = null,
    @SerialName("chunk_overlap") val chunkOverlap: Int? = null,
    @SerialName("chunk_strategy") val chunkStrategy: String? = null,
    @SerialName("entity_extraction_enabled") val entityExtractionEnabled: Boolean? = null,
    @SerialName("entity_extraction_strategy") val entityExtractionStrategy: String? = null,
    @SerialName("rerank_enabled") val rerankEnabled: Boolean? = null,
)

@Serializable
data class UpdateKnowledgeBaseRequest(
    val name: String? = null,
    val description: String? = null,
    val visibility: String? = null,
    @SerialName("chunk_size") val chunkSize: Int? = null,
    @SerialName("chunk_overlap") val chunkOverlap: Int? = null,
    @SerialName("chunk_strategy") val chunkStrategy: String? = null,
    val enabled: Boolean? = null,
    @SerialName("entity_extraction_enabled") val entityExtractionEnabled: Boolean? = null,
    @SerialName("entity_extraction_strategy") val entityExtractionStrategy: String? = null,
    @SerialName("rerank_enabled") val rerankEnabled: Boolean? = null,
)

@Serializable
data class KnowledgeBaseDocument(
    val id: String,
    @SerialName("knowledge_base_id") val knowledgeBaseId: String? = null,
    val title: String? = null,
    @SerialName("source_type") val sourceType: String? = null,
    @SerialName("mime_type") val mimeType: String? = null,
    val content: String? = null,
    val status: String? = null,
    @SerialName("chunks_count") val chunksCount: Int = 0,
    val tags: List<String> = emptyList(),
    @SerialName("created_at") val createdAt: String? = null,
    @SerialName("updated_at") val updatedAt: String? = null,
)

@Serializable
data class AddDocumentRequest(
    val title: String? = null,
    val content: String,
    @SerialName("source_type") val sourceType: String? = null,
    @SerialName("source_url") val sourceUrl: String? = null,
    @SerialName("mime_type") val mimeType: String? = null,
    val tags: List<String> = emptyList(),
    val metadata: Map<String, JsonElement>? = null,
)

@Serializable
data class KBSearchRequest(
    val query: String,
    val limit: Int? = null,
    val mode: String? = null,
    @SerialName("semantic_weight") val semanticWeight: Double? = null,
    val threshold: Double? = null,
    val rerank: Boolean? = null,
)

@Serializable
data class KBSearchResult(
    @SerialName("chunk_id") val chunkId: String,
    @SerialName("document_id") val documentId: String? = null,
    @SerialName("document_title") val documentTitle: String? = null,
    @SerialName("knowledge_base_name") val knowledgeBaseName: String? = null,
    val content: String,
    val similarity: Double,
    val metadata: Map<String, JsonElement>? = null,
)

@Serializable
data class Entity(
    val id: String,
    @SerialName("knowledge_base_id") val knowledgeBaseId: String? = null,
    @SerialName("entity_type") val entityType: String,
    val name: String,
    @SerialName("canonical_name") val canonicalName: String? = null,
    val aliases: List<String> = emptyList(),
)

@Serializable
data class EntityRelationship(
    val id: String,
    @SerialName("source_entity_id") val sourceEntityId: String,
    @SerialName("target_entity_id") val targetEntityId: String,
    @SerialName("relationship_type") val relationshipType: String,
    val direction: String? = null,
)

@Serializable
data class KnowledgeGraphData(
    val entities: List<Entity> = emptyList(),
    val relationships: List<EntityRelationship> = emptyList(),
)

@Serializable
data class KBPermissionGrant(
    val id: String,
    @SerialName("knowledge_base_id") val knowledgeBaseId: String? = null,
    @SerialName("document_id") val documentId: String? = null,
    @SerialName("user_id") val userId: String,
    val permission: String,
    @SerialName("granted_by") val grantedBy: String? = null,
    @SerialName("granted_at") val grantedAt: String? = null,
)

@Serializable
data class GrantKBPermissionRequest(
    @SerialName("user_id") val userId: String,
    val permission: String,
)

@Serializable
data class KBLinkedChatbot(
    val id: String,
    val name: String,
    val enabled: Boolean? = null,
    @SerialName("access_level") val accessLevel: String? = null,
)
