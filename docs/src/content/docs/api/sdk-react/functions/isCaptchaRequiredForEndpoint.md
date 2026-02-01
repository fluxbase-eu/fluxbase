---
editUrl: false
next: false
prev: false
title: "isCaptchaRequiredForEndpoint"
---

> **isCaptchaRequiredForEndpoint**(`config`, `endpoint`): `boolean`

Check if CAPTCHA is required for a specific endpoint

## Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `config` | [`CaptchaConfig`](/api/sdk-react/interfaces/captchaconfig/) \| `undefined` | CAPTCHA configuration from useCaptchaConfig |
| `endpoint` | `string` | The endpoint to check (e.g., 'signup', 'login', 'password_reset') |

## Returns

`boolean`

Whether CAPTCHA is required for this endpoint
