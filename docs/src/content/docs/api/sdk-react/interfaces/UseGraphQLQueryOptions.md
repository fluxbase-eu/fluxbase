---
editUrl: false
next: false
prev: false
title: "UseGraphQLQueryOptions"
---

Options for useGraphQLQuery hook

## Type Parameters

| Type Parameter |
| ------ |
| `T` |

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="enabled"></a> `enabled?` | `boolean` | Whether the query is enabled **Default** `true` |
| <a id="gctime"></a> `gcTime?` | `number` | Time in milliseconds after which inactive query data is garbage collected **Default** `5 minutes` |
| <a id="operationname"></a> `operationName?` | `string` | Operation name when the document contains multiple operations |
| <a id="refetchonwindowfocus"></a> `refetchOnWindowFocus?` | `boolean` | Whether to refetch on window focus **Default** `true` |
| <a id="requestoptions"></a> `requestOptions?` | [`GraphQLRequestOptions`](/api/sdk-react/interfaces/graphqlrequestoptions/) | Additional request options |
| <a id="select"></a> `select?` | (`data`) => `T` \| `undefined` | Transform function to process the response data |
| <a id="staletime"></a> `staleTime?` | `number` | Time in milliseconds after which the query is considered stale **Default** `0 (considered stale immediately)` |
| <a id="variables"></a> `variables?` | `Record`\<`string`, `unknown`\> | Variables to pass to the GraphQL query |
