---
editUrl: false
next: false
prev: false
title: "AIProvider"
---

AI provider configuration

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| `config` | `Record`\<`string`, `string`\> | - |
| `created_at` | `string` | - |
| `display_name` | `string` | - |
| `embedding_model` | `null` \| `string` | Embedding model for this provider. null means use provider-specific default |
| `enabled` | `boolean` | - |
| `from_config?` | `boolean` | True if provider was configured via environment variables or fluxbase.yaml |
| `id` | `string` | - |
| `is_default` | `boolean` | - |
| `name` | `string` | - |
| `provider_type` | [`AIProviderType`](/api/sdk/type-aliases/aiprovidertype/) | - |
| ~~`read_only?`~~ | `boolean` | :::caution[Deprecated] Use from_config instead ::: |
| `updated_at` | `string` | - |
| `use_for_embeddings` | `null` \| `boolean` | When true, this provider is explicitly used for embeddings. null means auto (follow default provider) |
