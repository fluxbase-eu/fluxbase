---
editUrl: false
next: false
prev: false
title: "AuthResponseData"
---

> **AuthResponseData** = `object`

Auth response with user and session (Supabase-compatible)

## Properties

| Property | Type |
| ------ | ------ |
| <a id="session"></a> `session` | [`AuthSession`](/api/sdk/interfaces/authsession/) \| `null` |
| <a id="user"></a> `user` | [`User`](/api/sdk/interfaces/user/) |
| <a id="weakpassword"></a> `weakPassword?` | [`WeakPassword`](/api/sdk/interfaces/weakpassword/) |
