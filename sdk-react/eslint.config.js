import js from '@eslint/js'
import tseslint from 'typescript-eslint'

// Flat config for the React SDK. Plain type-info-free typescript-eslint —
// eslint-plugin-react-hooks is intentionally NOT added (no new heavy deps
// for this package; add it + the plugin's recommended rules if the package
// starts carrying hook-heavy code). `bun run lint` lints src/ only.
export default tseslint.config(
  { ignores: ['dist', 'node_modules', 'coverage'] },
  {
    files: ['src/**/*.{ts,tsx}'],
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    rules: {
      // TODO(lint): ~250 violations across the hooks layer (options objects,
      // event payloads). Fix by typing the boundaries properly.
      '@typescript-eslint/no-explicit-any': 'warn',
      // TODO(lint): 6 violations (callback-typed options declared as
      // `Function`). Replace with specific signatures.
      '@typescript-eslint/no-unsafe-function-type': 'warn',
      // TODO(lint): ~30 violations, mostly unused test imports. Remove them.
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
