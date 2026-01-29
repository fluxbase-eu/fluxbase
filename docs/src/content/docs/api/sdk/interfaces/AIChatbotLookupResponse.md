---
editUrl: false
next: false
prev: false
title: "AIChatbotLookupResponse"
---

Response from chatbot lookup by name (smart namespace resolution)

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| `ambiguous` | `boolean` | True if multiple chatbots with this name exist in different namespaces |
| `chatbot?` | [`AIChatbotSummary`](/api/sdk/interfaces/aichatbotsummary/) | The chatbot if found (unique or resolved from default namespace) |
| `error?` | `string` | Error message if lookup failed |
| `namespaces?` | `string`[] | List of namespaces where the chatbot exists (when ambiguous) |
