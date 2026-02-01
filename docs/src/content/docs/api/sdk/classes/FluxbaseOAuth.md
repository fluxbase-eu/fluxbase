---
editUrl: false
next: false
prev: false
title: "FluxbaseOAuth"
---

OAuth Configuration Manager

Root manager providing access to OAuth provider and authentication settings management.

## Example

```typescript
const oauth = client.admin.oauth

// Manage OAuth providers
const providers = await oauth.providers.listProviders()

// Manage auth settings
const settings = await oauth.authSettings.get()
```

## Constructors

### Constructor

> **new FluxbaseOAuth**(`fetch`): `FluxbaseOAuth`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |

#### Returns

`FluxbaseOAuth`

## Properties

| Property | Modifier | Type |
| ------ | ------ | ------ |
| <a id="authsettings"></a> `authSettings` | `public` | [`AuthSettingsManager`](/api/sdk/classes/authsettingsmanager/) |
| <a id="providers"></a> `providers` | `public` | [`OAuthProviderManager`](/api/sdk/classes/oauthprovidermanager/) |
