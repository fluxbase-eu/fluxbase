import js from '@eslint/js'
import tseslint from 'typescript-eslint'

// Flat config for the Next.js SDK. Plain type-info-free typescript-eslint —
// the Next-specific rulesets (eslint-config-next) pull in React/compiler
// tooling this package does not depend on. `bun run lint` lints src/ only.
export default tseslint.config(
  { ignores: ['dist', 'node_modules', 'coverage'] },
  {
    files: ['src/**/*.{ts,tsx}'],
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    rules: {
      // create-server-client.ts lazily `require`s next/headers so the module
      // is testable without a Next runtime — that is deliberate, keep the
      // rule non-gating here.
      '@typescript-eslint/no-require-imports': 'warn',
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          args: 'all',
          argsIgnorePattern: '^_',
          caughtErrors: 'all',
          caughtErrorsIgnorePattern: '^_',
          destructuredArrayIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          ignoreRestSiblings: true,
        },
      ],
    },
  },
)
