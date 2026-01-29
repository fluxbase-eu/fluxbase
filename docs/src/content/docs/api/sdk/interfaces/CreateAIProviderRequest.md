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
| `config` | `Record`\<`string`, `string` \| `number` \| `boolean`\> | - |
| `display_name` | `string` | - |
| `embedding_model?` | `null` \| `string` | Embedding model for this provider. null or omit to use provider-specific default |
| `enabled?` | `boolean` | - |
| `is_default?` | `boolean` | - |
| `name` | `string` | - |
| `provider_type` | [`AIProviderType`](/api/sdk/type-aliases/aiprovidertype/) | - |
