---
editUrl: false
next: false
prev: false
title: "MigrationExecution"
---

Migration execution record (audit log)

## Properties

| Property | Type |
| ------ | ------ |
| <a id="action"></a> `action` | `"apply"` \| `"rollback"` |
| <a id="duration_ms"></a> `duration_ms?` | `number` |
| <a id="error_message"></a> `error_message?` | `string` |
| <a id="executed_at"></a> `executed_at` | `string` |
| <a id="executed_by"></a> `executed_by?` | `string` |
| <a id="id"></a> `id` | `string` |
| <a id="logs"></a> `logs?` | `string` |
| <a id="migration_id"></a> `migration_id` | `string` |
| <a id="status"></a> `status` | `"success"` \| `"failed"` |
