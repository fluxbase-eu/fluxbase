---
editUrl: false
next: false
prev: false
title: "UpdateEmailProviderSettingsRequest"
---

Request to update email provider settings

All fields are optional - only provided fields will be updated.
Secret fields (passwords, client keys) are only updated if provided.

## Properties

| Property | Type |
| ------ | ------ |
| <a id="enabled"></a> `enabled?` | `boolean` |
| <a id="from_address"></a> `from_address?` | `string` |
| <a id="from_name"></a> `from_name?` | `string` |
| <a id="mailgun_api_key"></a> `mailgun_api_key?` | `string` |
| <a id="mailgun_domain"></a> `mailgun_domain?` | `string` |
| <a id="provider"></a> `provider?` | `"smtp"` \| `"sendgrid"` \| `"mailgun"` \| `"ses"` |
| <a id="sendgrid_api_key"></a> `sendgrid_api_key?` | `string` |
| <a id="ses_access_key"></a> `ses_access_key?` | `string` |
| <a id="ses_region"></a> `ses_region?` | `string` |
| <a id="ses_secret_key"></a> `ses_secret_key?` | `string` |
| <a id="smtp_host"></a> `smtp_host?` | `string` |
| <a id="smtp_password"></a> `smtp_password?` | `string` |
| <a id="smtp_port"></a> `smtp_port?` | `number` |
| <a id="smtp_tls"></a> `smtp_tls?` | `boolean` |
| <a id="smtp_username"></a> `smtp_username?` | `string` |
