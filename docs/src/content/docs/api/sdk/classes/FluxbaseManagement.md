---
editUrl: false
next: false
prev: false
title: "FluxbaseManagement"
---

Management client for client keys, webhooks, and invitations

## Constructors

### Constructor

> **new FluxbaseManagement**(`fetch`): `FluxbaseManagement`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |

#### Returns

`FluxbaseManagement`

## Properties

| Property | Modifier | Type | Description |
| ------ | ------ | ------ | ------ |
| <a id="apikeys"></a> ~~`apiKeys`~~ | `public` | [`ClientKeysManager`](/api/sdk/classes/clientkeysmanager/) | :::caution[Deprecated] Use clientKeys instead ::: |
| <a id="clientkeys"></a> `clientKeys` | `public` | [`ClientKeysManager`](/api/sdk/classes/clientkeysmanager/) | Client Keys management |
| <a id="invitations"></a> `invitations` | `public` | [`InvitationsManager`](/api/sdk/classes/invitationsmanager/) | Invitations management |
| <a id="webhooks"></a> `webhooks` | `public` | [`WebhooksManager`](/api/sdk/classes/webhooksmanager/) | Webhooks management |
