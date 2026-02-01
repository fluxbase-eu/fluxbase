---
editUrl: false
next: false
prev: false
title: "UpsertOptions"
---

Options for upsert operations (Supabase-compatible)

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="defaulttonull"></a> `defaultToNull?` | `boolean` | If true, missing columns default to null instead of using existing values **Default** `false` |
| <a id="ignoreduplicates"></a> `ignoreDuplicates?` | `boolean` | If true, duplicate rows are ignored (not upserted) **Default** `false` |
| <a id="onconflict"></a> `onConflict?` | `string` | Comma-separated columns to use for conflict resolution **Examples** `'email'` `'user_id,tenant_id'` |
