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
| <a id="local_tokens_revoked"></a> `local_tokens_revoked` | `boolean` | Whether local JWT tokens were revoked |
| <a id="provider"></a> `provider` | `string` | OAuth provider name |
| <a id="provider_token_revoked"></a> `provider_token_revoked` | `boolean` | Whether the token was revoked at the OAuth provider |
| <a id="redirect_url"></a> `redirect_url?` | `string` | URL to redirect to for OIDC logout (if requires_redirect is true) |
| <a id="requires_redirect"></a> `requires_redirect?` | `boolean` | Whether the user should be redirected to the provider's logout page |
| <a id="warning"></a> `warning?` | `string` | Warning message if something failed but logout still proceeded |
