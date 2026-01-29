---
editUrl: false
next: false
prev: false
title: "OAuthLogoutResponse"
---

Response from OAuth logout endpoint

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| `local_tokens_revoked` | `boolean` | Whether local JWT tokens were revoked |
| `provider` | `string` | OAuth provider name |
| `provider_token_revoked` | `boolean` | Whether the token was revoked at the OAuth provider |
| `redirect_url?` | `string` | URL to redirect to for OIDC logout (if requires_redirect is true) |
| `requires_redirect?` | `boolean` | Whether the user should be redirected to the provider's logout page |
| `warning?` | `string` | Warning message if something failed but logout still proceeded |
