package io.github.nimbleflux.fluxbase.knowledgebase

import io.github.nimbleflux.fluxbase.FluxbaseResponse
import io.github.nimbleflux.fluxbase.core.FluxbaseHttpClient
import io.github.nimbleflux.fluxbase.fluxbaseResponse
/**
 * Knowledge base module — user-scope CRUD, documents, hybrid search, and the
 * knowledge graph. Mirrors the TS SDK's `FluxbaseKnowledgeBase`
 * (`sdk/src/knowledge-base.ts`).
 */
class FluxbaseKnowledgeBase(private val http: FluxbaseHttpClient) {

    private fun encode(s: String): String = java.net.URLEncoder.encode(s, "UTF-8")

    // ------------------------------------------------------------------
    // Knowledge base CRUD
    // ------------------------------------------------------------------

    /** GET `/api/v1/ai/knowledge-bases` — list KBs the user can access. */
    suspend fun list(namespace: String? = null): FluxbaseResponse<List<KnowledgeBaseSummary>> =
        fluxbaseResponse {
            val query = namespace?.let { "?namespace=${encode(it)}" } ?: ""
            http.get("/api/v1/ai/knowledge-bases$query")
        }

    /** POST `/api/v1/ai/knowledge-bases` — create a KB owned by the caller. */
    suspend fun create(request: CreateKnowledgeBaseRequest): FluxbaseResponse<KnowledgeBase> =
        fluxbaseResponse {
            http.post("/api/v1/ai/knowledge-bases", request)
        }

    /** GET `/api/v1/ai/knowledge-bases/:id` — get a KB (requires access). */
    suspend fun get(kbId: String): FluxbaseResponse<KnowledgeBase> = fluxbaseResponse {
        http.get("/api/v1/ai/knowledge-bases/$kbId")
    }

    /** PATCH `/api/v1/ai/knowledge-bases/:id` — update (editor permission). */
    suspend fun update(kbId: String, request: UpdateKnowledgeBaseRequest): FluxbaseResponse<KnowledgeBase> =
        fluxbaseResponse {
            http.patch("/api/v1/ai/knowledge-bases/$kbId", request)
        }

    /** DELETE `/api/v1/ai/knowledge-bases/:id` — delete (owner only). */
    suspend fun delete(kbId: String): FluxbaseResponse<Unit> = fluxbaseResponse {
        http.delete("/api/v1/ai/knowledge-bases/$kbId")
        Unit
    }

    // ------------------------------------------------------------------
    // Sharing & permissions
    // ------------------------------------------------------------------

    /** POST `/api/v1/ai/knowledge-bases/:id/share` — grant a user permission (owner). */
    suspend fun share(kbId: String, request: GrantKBPermissionRequest): FluxbaseResponse<KBPermissionGrant> =
        fluxbaseResponse {
            http.post("/api/v1/ai/knowledge-bases/$kbId/share", request)
        }

    /** GET `/api/v1/ai/knowledge-bases/:id/permissions` — list grants (owner). */
    suspend fun listPermissions(kbId: String): FluxbaseResponse<List<KBPermissionGrant>> =
        fluxbaseResponse {
            http.get("/api/v1/ai/knowledge-bases/$kbId/permissions")
        }

    /** DELETE `/api/v1/ai/knowledge-bases/:id/permissions/:userId` — revoke (owner). */
    suspend fun revokePermission(kbId: String, userId: String): FluxbaseResponse<Unit> =
        fluxbaseResponse {
            http.delete("/api/v1/ai/knowledge-bases/$kbId/permissions/${encode(userId)}")
            Unit
        }

    // ------------------------------------------------------------------
    // Documents
    // ------------------------------------------------------------------

    /** GET `/api/v1/ai/knowledge-bases/:id/documents` — list documents. */
    suspend fun listDocuments(kbId: String): FluxbaseResponse<List<KnowledgeBaseDocument>> =
        fluxbaseResponse {
            http.get("/api/v1/ai/knowledge-bases/$kbId/documents")
        }

    /** GET `/api/v1/ai/knowledge-bases/:id/documents/:docId` — get a document. */
    suspend fun getDocument(kbId: String, documentId: String): FluxbaseResponse<KnowledgeBaseDocument> =
        fluxbaseResponse {
            http.get("/api/v1/ai/knowledge-bases/$kbId/documents/$documentId")
        }

    /** POST `/api/v1/ai/knowledge-bases/:id/documents` — add a text document. */
    suspend fun addDocument(kbId: String, request: AddDocumentRequest): FluxbaseResponse<KnowledgeBaseDocument> =
        fluxbaseResponse {
            http.post("/api/v1/ai/knowledge-bases/$kbId/documents", request)
        }

    /**
     * POST `/api/v1/ai/knowledge-bases/:id/documents/upload` — upload a file
     * as a multipart form document. The multipart body is assembled manually
     * (the HTTP transport has no multipart helper); error statuses throw
     * inside the transport and are converted by [fluxbaseResponse].
     */
    suspend fun uploadDocument(
        kbId: String,
        fileName: String,
        content: ByteArray,
        mimeType: String,
        metadata: Map<String, String>? = null,
    ): FluxbaseResponse<KnowledgeBaseDocument> = fluxbaseResponse {
        val boundary = "----fluxbase" + java.util.UUID.randomUUID().toString().replace("-", "")
        val body = buildMultipartBody(
            boundary = boundary,
            fileField = "file",
            fileName = fileName,
            fileContent = content,
            fileContentType = mimeType,
            fields = metadata ?: emptyMap(),
        )
        val headers = mapOf("Content-Type" to "multipart/form-data; boundary=$boundary")
        val response = http.transport.request(
            io.github.nimbleflux.fluxbase.core.HttpMethod.POST,
            "/api/v1/ai/knowledge-bases/$kbId/documents/upload",
            body,
            http.defaultHeaders.toMap() + headers,
        )
        io.github.nimbleflux.fluxbase.core.FluxbaseHttpClient.defaultJson
            .decodeFromString(KnowledgeBaseDocument.serializer(), response.body)
    }

    /** Assembles a multipart/form-data body with optional text fields plus one file part. */
    private fun buildMultipartBody(
        boundary: String,
        fileField: String,
        fileName: String,
        fileContent: ByteArray,
        fileContentType: String,
        fields: Map<String, String>,
    ): ByteArray {
        val charset = Charsets.UTF_8
        val out = java.io.ByteArrayOutputStream()
        fun write(s: String) = out.write(s.toByteArray(charset))

        for ((name, value) in fields) {
            write("--$boundary\r\n")
            write("Content-Disposition: form-data; name=\"${name}\"\r\n\r\n")
            write("$value\r\n")
        }
        write("--$boundary\r\n")
        write("Content-Disposition: form-data; name=\"$fileField\"; filename=\"$fileName\"\r\n")
        write("Content-Type: $fileContentType\r\n\r\n")
        out.write(fileContent)
        write("\r\n")
        write("--$boundary--\r\n")
        return out.toByteArray()
    }

    /** DELETE `/api/v1/ai/knowledge-bases/:id/documents/:docId` — delete a document. */
    suspend fun deleteDocument(kbId: String, documentId: String): FluxbaseResponse<Unit> =
        fluxbaseResponse {
            http.delete("/api/v1/ai/knowledge-bases/$kbId/documents/$documentId")
            Unit
        }

    // ------------------------------------------------------------------
    // Search
    // ------------------------------------------------------------------

    /** POST `/api/v1/ai/knowledge-bases/:id/search` — hybrid search. */
    suspend fun search(kbId: String, request: KBSearchRequest): FluxbaseResponse<List<KBSearchResult>> =
        fluxbaseResponse {
            http.post("/api/v1/ai/knowledge-bases/$kbId/search", request)
        }

    // ------------------------------------------------------------------
    // Entities & knowledge graph
    // ------------------------------------------------------------------

    /** GET `/api/v1/ai/knowledge-bases/:id/entities` — list entities. */
    suspend fun listEntities(kbId: String, type: String? = null): FluxbaseResponse<List<Entity>> =
        fluxbaseResponse {
            val query = type?.let { "?type=${encode(it)}" } ?: ""
            http.get("/api/v1/ai/knowledge-bases/$kbId/entities$query")
        }

    /** GET `/api/v1/ai/knowledge-bases/:id/entities/search?q=…` — search entities by name. */
    suspend fun searchEntities(kbId: String, query: String): FluxbaseResponse<List<Entity>> =
        fluxbaseResponse {
            http.get("/api/v1/ai/knowledge-bases/$kbId/entities/search?q=${encode(query)}")
        }

    /** GET `/api/v1/ai/knowledge-bases/:id/entities/:entityId/relationships`. */
    suspend fun getEntityRelationships(kbId: String, entityId: String): FluxbaseResponse<List<EntityRelationship>> =
        fluxbaseResponse {
            http.get("/api/v1/ai/knowledge-bases/$kbId/entities/${encode(entityId)}/relationships")
        }

    /** GET `/api/v1/ai/knowledge-bases/:id/graph` — full knowledge graph. */
    suspend fun getKnowledgeGraph(kbId: String): FluxbaseResponse<KnowledgeGraphData> =
        fluxbaseResponse {
            http.get("/api/v1/ai/knowledge-bases/$kbId/graph")
        }

    // ------------------------------------------------------------------
    // Chatbot links
    // ------------------------------------------------------------------

    /** GET `/api/v1/ai/knowledge-bases/:id/chatbots` — chatbots linked to this KB. */
    suspend fun listChatbots(kbId: String): FluxbaseResponse<List<KBLinkedChatbot>> =
        fluxbaseResponse {
            http.get("/api/v1/ai/knowledge-bases/$kbId/chatbots")
        }
}
