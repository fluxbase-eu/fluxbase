---
editUrl: false
next: false
prev: false
title: "AuthResponse"
---

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="access_token"></a> `access_token` | `string` | - |
| <a id="expires_in"></a> `expires_in` | `number` | - |
| <a id="message"></a> `message?` | `string` | Human-readable message (present on requires_email_verification responses) |
| <a id="refresh_token"></a> `refresh_token` | `string` | - |
| <a id="requires_email_verification"></a> `requires_email_verification?` | `boolean` | Signups requiring email verification return no tokens and this flag |
| <a id="trust_token"></a> `trust_token?` | `string` | Issued after CAPTCHA verification for subsequent trust checks |
| <a id="user"></a> `user` | [`User`](/api/sdk/interfaces/user/) | - |
