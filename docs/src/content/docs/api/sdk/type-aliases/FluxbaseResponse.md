---
editUrl: false
next: false
prev: false
title: "FluxbaseResponse"
---

> **FluxbaseResponse**\<`T`\> = \{ `data`: `T`; `error`: `null`; \} \| \{ `data`: `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/); \}

Base Fluxbase response type (Supabase-compatible)
Returns either `{ data, error: null }` on success or `{ data: null, error }` on failure

The failure branch carries a [FluxbaseError](/api/sdk/interfaces/fluxbaseerror/): the error instance the
SDK's fetch layer actually produces, hydrated with `status`, `code` and
`details` from the server envelope.

## Type Parameters

| Type Parameter |
| ------ |
| `T` |
