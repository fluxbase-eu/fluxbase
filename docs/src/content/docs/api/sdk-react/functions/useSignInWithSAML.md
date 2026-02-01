---
editUrl: false
next: false
prev: false
title: "useSignInWithSAML"
---

> **useSignInWithSAML**(): `UseMutationResult`\<`DataResponse`\<\{ `provider`: `string`; `url`: `string`; \}\>, `Error`, \{ `options?`: [`SAMLLoginOptions`](/api/sdk-react/interfaces/samlloginoptions/); `provider`: `string`; \}, `unknown`\>

Hook to initiate SAML login (redirects to IdP)

This hook returns a mutation that when called, redirects the user to the
SAML Identity Provider for authentication.

## Returns

`UseMutationResult`\<`DataResponse`\<\{ `provider`: `string`; `url`: `string`; \}\>, `Error`, \{ `options?`: [`SAMLLoginOptions`](/api/sdk-react/interfaces/samlloginoptions/); `provider`: `string`; \}, `unknown`\>

## Example

```tsx
function SAMLLoginButton() {
  const signInWithSAML = useSignInWithSAML()

  return (
    <button
      onClick={() => signInWithSAML.mutate({ provider: 'okta' })}
      disabled={signInWithSAML.isPending}
    >
      {signInWithSAML.isPending ? 'Redirecting...' : 'Sign in with Okta'}
    </button>
  )
}
```
