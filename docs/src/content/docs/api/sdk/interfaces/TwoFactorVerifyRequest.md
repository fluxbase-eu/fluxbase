---
editUrl: false
next: false
prev: false
title: "TwoFactorVerifyRequest"
---

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="code"></a> `code` | `string` | - |
| <a id="mfa_token"></a> `mfa_token` | `string` | Short-lived signed token returned by signIn when requires_2fa is true |
| <a id="user_id"></a> `user_id?` | `string` | User ID (optional; must match the token's user when provided) |
