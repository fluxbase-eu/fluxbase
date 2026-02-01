---
editUrl: false
next: false
prev: false
title: "CaptchaConfig"
---

Public CAPTCHA configuration returned from the server
Used by clients to know which CAPTCHA provider to load

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="cap_server_url"></a> `cap_server_url?` | `string` | Cap server URL - only present when provider is 'cap' |
| <a id="enabled"></a> `enabled` | `boolean` | Whether CAPTCHA is enabled |
| <a id="endpoints"></a> `endpoints?` | `string`[] | Endpoints that require CAPTCHA verification |
| <a id="provider"></a> `provider?` | [`CaptchaProvider`](/api/sdk/type-aliases/captchaprovider/) | CAPTCHA provider name |
| <a id="site_key"></a> `site_key?` | `string` | Public site key for the CAPTCHA widget (hcaptcha, recaptcha, turnstile) |
