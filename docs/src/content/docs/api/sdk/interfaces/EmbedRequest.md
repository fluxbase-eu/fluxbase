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
| `model?` | `string` | Embedding model to use (defaults to configured model) |
| `provider?` | `string` | Provider ID to use for embedding (admin-only, defaults to configured embedding provider) |
| `text?` | `string` | Text to embed (single) |
| `texts?` | `string`[] | Multiple texts to embed |
