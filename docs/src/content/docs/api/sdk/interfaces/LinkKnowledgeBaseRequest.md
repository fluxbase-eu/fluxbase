---
editUrl: false
next: false
prev: false
title: "LinkKnowledgeBaseRequest"
---

Request to link a knowledge base to a chatbot

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="access_level"></a> `access_level?` | `"full"` \| `"filtered"` \| `"tiered"` | Access level: "full" exposes all chunks to the chatbot, "filtered" scopes retrieval to the calling user (default when omitted), "tiered" uses priority-based routing. |
| <a id="knowledge_base_id"></a> `knowledge_base_id` | `string` | - |
| <a id="max_chunks"></a> `max_chunks?` | `number` | - |
| <a id="priority"></a> `priority?` | `number` | - |
| <a id="similarity_threshold"></a> `similarity_threshold?` | `number` | - |
