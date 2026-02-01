---
editUrl: false
next: false
prev: false
title: "EmbedRequest"
---

Request for vector embedding generation

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="model"></a> `model?` | `string` | Embedding model to use (defaults to configured model) |
| <a id="provider"></a> `provider?` | `string` | Provider ID to use for embedding (admin-only, defaults to configured embedding provider) |
| <a id="text"></a> `text?` | `string` | Text to embed (single) |
| <a id="texts"></a> `texts?` | `string`[] | Multiple texts to embed |
