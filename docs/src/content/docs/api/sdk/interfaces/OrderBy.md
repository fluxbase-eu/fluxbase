---
editUrl: false
next: false
prev: false
title: "OrderBy"
---

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="column"></a> `column` | `string` | - |
| <a id="direction"></a> `direction` | [`OrderDirection`](/api/sdk/type-aliases/orderdirection/) | - |
| <a id="nulls"></a> `nulls?` | `"first"` \| `"last"` | - |
| <a id="vectorop"></a> `vectorOp?` | `"vec_l2"` \| `"vec_cos"` \| `"vec_ip"` | Vector operator for similarity ordering (vec_l2, vec_cos, vec_ip) |
| <a id="vectorvalue"></a> `vectorValue?` | `number`[] | Vector value for similarity ordering |
