---
editUrl: false
next: false
prev: false
title: "CreateAIProviderRequest"
---

Request to create an AI provider
Note: config values can be strings, numbers, or booleans - they will be converted to strings automatically

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="config"></a> `config` | `Record`\<`string`, `string` \| `number` \| `boolean`\> | - |
| <a id="display_name"></a> `display_name` | `string` | - |
| <a id="embedding_model"></a> `embedding_model?` | `string` \| `null` | Embedding model for this provider. null or omit to use provider-specific default |
| <a id="enabled"></a> `enabled?` | `boolean` | - |
| <a id="is_default"></a> `is_default?` | `boolean` | - |
| <a id="name"></a> `name` | `string` | - |
| <a id="provider_type"></a> `provider_type` | [`AIProviderType`](/api/sdk/type-aliases/aiprovidertype/) | - |
