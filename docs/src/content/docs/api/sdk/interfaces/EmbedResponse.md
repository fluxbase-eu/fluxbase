---
editUrl: false
next: false
prev: false
title: "EmbedResponse"
---

Response from vector embedding generation

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="dimensions"></a> `dimensions` | `number` | Dimensions of the embeddings |
| <a id="embeddings"></a> `embeddings` | `number`[][] | Generated embeddings (one per input text) |
| <a id="model"></a> `model` | `string` | Model used for embedding |
| <a id="usage"></a> `usage?` | `object` | Token usage information |
| `usage.prompt_tokens` | `number` | - |
| `usage.total_tokens` | `number` | - |
