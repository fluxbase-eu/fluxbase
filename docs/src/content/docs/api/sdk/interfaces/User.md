---
editUrl: false
next: false
prev: false
title: "User"
---

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="app_metadata"></a> `app_metadata?` | `Record`\<`string`, `unknown`\> \| `null` | Application/admin-only metadata (auth.users.app_metadata) |
| <a id="created_at"></a> `created_at` | `string` | - |
| <a id="email"></a> `email` | `string` | - |
| <a id="email_verified"></a> `email_verified` | `boolean` | - |
| <a id="id"></a> `id` | `string` | - |
| <a id="metadata"></a> ~~`metadata?`~~ | `Record`\<`string`, `unknown`\> \| `null` | :::caution[Deprecated] Server sends `user_metadata` / `app_metadata`; kept for backwards compatibility ::: |
| <a id="role"></a> ~~`role?`~~ | `string` | :::caution[Deprecated] Server wire field is optional (omitempty) and admin-managed ::: |
| <a id="updated_at"></a> `updated_at` | `string` | - |
| <a id="user_metadata"></a> `user_metadata?` | `Record`\<`string`, `unknown`\> \| `null` | User-editable metadata (auth.users.user_metadata) |
