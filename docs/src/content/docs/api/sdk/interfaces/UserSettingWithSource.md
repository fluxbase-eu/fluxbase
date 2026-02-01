---
editUrl: false
next: false
prev: false
title: "UserSettingWithSource"
---

A setting with source information (user or system)
Returned when fetching a setting with user -> system fallback

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="key"></a> `key` | `string` | - |
| <a id="source"></a> `source` | `"user"` \| `"system"` | Where the value came from: "user" = user's own setting, "system" = system default |
| <a id="value"></a> `value` | `Record`\<`string`, `unknown`\> | - |
