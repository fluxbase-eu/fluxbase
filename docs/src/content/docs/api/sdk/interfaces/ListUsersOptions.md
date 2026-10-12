---
editUrl: false
next: false
prev: false
title: "ListUsersOptions"
---

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="exclude_admins"></a> `exclude_admins?` | `boolean` | - |
| <a id="limit"></a> `limit?` | `number` | - |
| <a id="offset"></a> `offset?` | `number` | - |
| <a id="search"></a> `search?` | `string` | - |
| <a id="signal"></a> `signal?` | `AbortSignal` \| `null` | AbortSignal for caller-requested cancellation of the request |
| <a id="type"></a> `type?` | `"app"` \| `"platform"` | Which user table to query. Only "platform" is recognized by the server; any other value defaults to "app". |
