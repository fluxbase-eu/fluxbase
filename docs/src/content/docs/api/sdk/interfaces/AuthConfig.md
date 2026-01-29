---
editUrl: false
next: false
prev: false
title: "AuthConfig"
---

Comprehensive authentication configuration
Returns all public auth settings from the server

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| `captcha` | `null` \| [`CaptchaConfig`](/api/sdk/interfaces/captchaconfig/) | CAPTCHA configuration |
| `magic_link_enabled` | `boolean` | Whether magic link authentication is enabled |
| `mfa_available` | `boolean` | Whether MFA/2FA is available (always true, users opt-in) |
| `oauth_providers` | [`OAuthProviderPublic`](/api/sdk/interfaces/oauthproviderpublic/)[] | Available OAuth providers for authentication |
| `password_login_enabled` | `boolean` | Whether password login is enabled for app users |
| `password_min_length` | `number` | Minimum password length requirement |
| `password_require_lowercase` | `boolean` | Whether passwords must contain lowercase letters |
| `password_require_number` | `boolean` | Whether passwords must contain numbers |
| `password_require_special` | `boolean` | Whether passwords must contain special characters |
| `password_require_uppercase` | `boolean` | Whether passwords must contain uppercase letters |
| `require_email_verification` | `boolean` | Whether email verification is required after signup |
| `saml_providers` | [`SAMLProvider`](/api/sdk/interfaces/samlprovider/)[] | Available SAML providers for enterprise SSO |
| `signup_enabled` | `boolean` | Whether user signup is enabled |
