---
editUrl: false
next: false
prev: false
title: "useAdminAIConversationMessages"
---

> **useAdminAIConversationMessages**(`conversationId`, `query?`): `UseQueryResult`\<`AdminMessagesResponse` \| `null`, `Error`\>

Hook to get all messages of a conversation (admin)

## Parameters

| Parameter | Type |
| ------ | ------ |
| `conversationId` | `string` \| `null` |
| `query` | \{ `limit?`: `number`; `offset?`: `number`; \} |
| `query.limit?` | `number` |
| `query.offset?` | `number` |

## Returns

`UseQueryResult`\<`AdminMessagesResponse` \| `null`, `Error`\>
