---
editUrl: false
next: false
prev: false
title: "EnrichedUser"
---

User with enriched metadata as returned by the admin user management
endpoints. Mirrors the Go `auth.EnrichedUser` struct.

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="active_sessions"></a> `active_sessions` | `number` | - |
| <a id="app_metadata"></a> `app_metadata?` | `Record`\<`string`, `unknown`\> \| `null` | - |
| <a id="created_at"></a> `created_at` | `string` | - |
| <a id="email"></a> `email` | `string` | - |
| <a id="email_verified"></a> `email_verified` | `boolean` | - |
| <a id="id"></a> `id` | `string` | - |
| <a id="is_locked"></a> `is_locked` | `boolean` | - |
| <a id="last_sign_in"></a> `last_sign_in` | `string` \| `null` | Timestamp of the last sign-in, or null if the user never signed in |
| <a id="provider"></a> `provider` | `string` | "email", "invite_pending", or "magic_link" |
| <a id="role"></a> `role` | `string` | - |
| <a id="tenant_assignments"></a> `tenant_assignments?` | `TenantAssignment`[] | - |
| <a id="updated_at"></a> `updated_at` | `string` | - |
| <a id="user_metadata"></a> `user_metadata?` | `Record`\<`string`, `unknown`\> \| `null` | - |
