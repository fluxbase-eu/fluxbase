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
| `key` | `string` | - |
| `source` | `"user"` \| `"system"` | Where the value came from: "user" = user's own setting, "system" = system default |
| `value` | `Record`\<`string`, `unknown`\> | - |
