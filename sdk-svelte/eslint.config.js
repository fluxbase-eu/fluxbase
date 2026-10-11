import js from '@eslint/js'
import tseslint from 'typescript-eslint'

// Flat config for the Svelte SDK. Plain type-info-free typescript-eslint —
// this package ships stores/helpers written in plain .ts (no .svelte files),
// so eslint-plugin-svelte is not needed. `bun run lint` lints src/ only.
export default tseslint.config(
  { ignores: ['dist', 'node_modules', 'coverage', 'examples'] },
  {
    files: ['src/**/*.ts'],
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    rules: {
      // TODO(lint): ~45 violations in store/browser shims. Fix by typing the
      // boundaries properly.
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': [
        'warn',
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
