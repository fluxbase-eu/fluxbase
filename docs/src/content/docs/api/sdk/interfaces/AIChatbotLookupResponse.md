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
| <a id="ambiguous"></a> `ambiguous` | `boolean` | True if multiple chatbots with this name exist in different namespaces |
| <a id="chatbot"></a> `chatbot?` | [`AIChatbotSummary`](/api/sdk/interfaces/aichatbotsummary/) | The chatbot if found (unique or resolved from default namespace) |
| <a id="error"></a> `error?` | `string` | Error message if lookup failed |
| <a id="namespaces"></a> `namespaces?` | `string`[] | List of namespaces where the chatbot exists (when ambiguous) |
