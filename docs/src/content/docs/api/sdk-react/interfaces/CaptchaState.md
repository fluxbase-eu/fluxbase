---
editUrl: false
next: false
prev: false
title: "CaptchaState"
---

CAPTCHA widget state for managing token generation

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="error"></a> `error` | `Error` \| `null` | Any error that occurred |
| <a id="execute"></a> `execute` | () => `Promise`\<`string`\> | Execute/trigger the CAPTCHA (for invisible CAPTCHA like reCAPTCHA v3) |
| <a id="isloading"></a> `isLoading` | `boolean` | Whether a token is being generated |
| <a id="isready"></a> `isReady` | `boolean` | Whether the CAPTCHA widget is ready |
| <a id="onerror"></a> `onError` | (`error`) => `void` | Callback to be called when CAPTCHA errors |
| <a id="onexpire"></a> `onExpire` | () => `void` | Callback to be called when CAPTCHA expires |
| <a id="onverify"></a> `onVerify` | (`token`) => `void` | Callback to be called when CAPTCHA is verified |
| <a id="reset"></a> `reset` | () => `void` | Reset the CAPTCHA widget |
| <a id="token"></a> `token` | `string` \| `null` | Current CAPTCHA token (null until solved) |
