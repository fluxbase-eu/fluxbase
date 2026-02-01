---
editUrl: false
next: false
prev: false
title: "isAuthSuccess"
---

> **isAuthSuccess**(`response`): `response is { data: AuthResponseData; error: null }`

Type guard to check if an auth response is successful

## Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `response` | [`FluxbaseAuthResponse`](/api/sdk/type-aliases/fluxbaseauthresponse/) | The auth response to check |

## Returns

`response is { data: AuthResponseData; error: null }`

true if the auth operation succeeded
