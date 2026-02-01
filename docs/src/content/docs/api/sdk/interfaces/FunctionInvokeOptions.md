---
editUrl: false
next: false
prev: false
title: "FunctionInvokeOptions"
---

Options for invoking an edge function

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="body"></a> `body?` | `unknown` | Request body to send to the function |
| <a id="headers"></a> `headers?` | `Record`\<`string`, `string`\> | Custom headers to include in the request |
| <a id="method"></a> `method?` | `"GET"` \| `"POST"` \| `"PUT"` \| `"PATCH"` \| `"DELETE"` | HTTP method to use **Default** `'POST'` |
| <a id="namespace"></a> `namespace?` | `string` | Namespace of the function to invoke If not provided, the first function with the given name is used (alphabetically by namespace) |
