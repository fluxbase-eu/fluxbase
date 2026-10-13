package io.github.nimbleflux.fluxbase.core

import kotlinx.serialization.InternalSerializationApi
import kotlinx.serialization.KSerializer
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.encodeToJsonElement
import kotlinx.serialization.serializerOrNull

/**
 * Convert an arbitrary Kotlin value to a [JsonElement]. Handles Maps, Lists,
 * primitives, and passes pre-built [JsonElement] values through untouched.
 * Mirrors how the TS SDK passes plain JS objects through `JSON.stringify`.
 *
 * Shared by every body-encoding call site — [KtorHttpTransport] for request
 * bodies, [io.github.nimbleflux.fluxbase.rpc.FluxbaseRpc.invoke] for RPC params,
 * GraphQL variables, and realtime broadcast payloads — so all of them encode
 * `@Serializable` dataclasses via their generated serializer. Types without a
 * serializer fall back to `toString()` as a JSON string (this fallback
 * previously corrupted lists/maps into stringified Kotlin output).
 *
 * `@PublishedApi` because the public inline `FluxbaseGraphQL.query` calls it.
 */
@OptIn(InternalSerializationApi::class)
@PublishedApi
internal fun encodeToJsonElement(json: Json, value: Any?): JsonElement = when (value) {
    null -> JsonNull
    is JsonElement -> value
    is String -> JsonPrimitive(value)
    is Number -> JsonPrimitive(value)
    is Boolean -> JsonPrimitive(value)
    is Map<*, *> -> buildJsonObject {
        value.forEach { (k, v) -> put(k.toString(), encodeToJsonElement(json, v)) }
    }
    is List<*> -> buildJsonArray {
        value.forEach { add(encodeToJsonElement(json, it)) }
    }
    else -> {
        @Suppress("UNCHECKED_CAST")
        val serializer = value::class.serializerOrNull() as? KSerializer<Any>
        if (serializer != null) {
            json.encodeToJsonElement(serializer, value)
        } else {
            JsonPrimitive(value.toString())
        }
    }
}
