---
editUrl: false
next: false
prev: false
title: "EmailSettings"
---

Email configuration settings

## Properties

| Property | Type |
| ------ | ------ |
| <a id="enabled"></a> `enabled` | `boolean` |
| <a id="from_address"></a> `from_address?` | `string` |
| <a id="from_name"></a> `from_name?` | `string` |
| <a id="mailgun"></a> `mailgun?` | [`MailgunSettings`](/api/sdk/interfaces/mailgunsettings/) |
| <a id="provider"></a> `provider` | `"smtp"` \| `"sendgrid"` \| `"mailgun"` \| `"ses"` |
| <a id="reply_to_address"></a> `reply_to_address?` | `string` |
| <a id="sendgrid"></a> `sendgrid?` | [`SendGridSettings`](/api/sdk/interfaces/sendgridsettings/) |
| <a id="ses"></a> `ses?` | [`SESSettings`](/api/sdk/interfaces/sessettings/) |
| <a id="smtp"></a> `smtp?` | [`SMTPSettings`](/api/sdk/interfaces/smtpsettings/) |
