import js from '@eslint/js'
import tseslint from 'typescript-eslint'

// Flat config for the Vue SDK. Plain type-info-free typescript-eslint —
// this package ships composables written in plain .ts (no .vue SFCs), so
// vue-eslint-parser is not needed. `bun run lint` lints src/ only.
export default tseslint.config(
  { ignores: ['dist', 'node_modules', 'coverage'] },
  {
    files: ['src/**/*.ts'],
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    rules: {
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
