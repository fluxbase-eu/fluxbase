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
| <a id="config"></a> `config` | `Record`\<`string`, `string`\> | - |
| <a id="created_at"></a> `created_at` | `string` | - |
| <a id="display_name"></a> `display_name` | `string` | - |
| <a id="embedding_model"></a> `embedding_model` | `string` \| `null` | Embedding model for this provider. null means use provider-specific default |
| <a id="enabled"></a> `enabled` | `boolean` | - |
| <a id="from_config"></a> `from_config?` | `boolean` | True if provider was configured via environment variables or fluxbase.yaml |
| <a id="id"></a> `id` | `string` | - |
| <a id="is_default"></a> `is_default` | `boolean` | - |
| <a id="name"></a> `name` | `string` | - |
| <a id="provider_type"></a> `provider_type` | [`AIProviderType`](/api/sdk/type-aliases/aiprovidertype/) | - |
| <a id="read_only"></a> ~~`read_only?`~~ | `boolean` | :::caution[Deprecated] Use from_config instead ::: |
| <a id="updated_at"></a> `updated_at` | `string` | - |
| <a id="use_for_embeddings"></a> `use_for_embeddings` | `boolean` \| `null` | When true, this provider is explicitly used for embeddings. null means auto (follow default provider) |
