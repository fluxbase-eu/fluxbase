---
editUrl: false
next: false
prev: false
title: "UseGraphQLMutationOptions"
---

Options for useGraphQLMutation hook

## Type Parameters

| Type Parameter |
| ------ |
| `T` |
| `V` |

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="invalidatequeries"></a> `invalidateQueries?` | `string`[] | Query keys to invalidate on success |
| <a id="onerror"></a> `onError?` | (`error`, `variables`) => `void` | Callback when mutation fails |
| <a id="onsuccess"></a> `onSuccess?` | (`data`, `variables`) => `void` | Callback when mutation succeeds |
| <a id="operationname"></a> `operationName?` | `string` | Operation name when the document contains multiple operations |
| <a id="requestoptions"></a> `requestOptions?` | [`GraphQLRequestOptions`](/api/sdk-react/interfaces/graphqlrequestoptions/) | Additional request options |
