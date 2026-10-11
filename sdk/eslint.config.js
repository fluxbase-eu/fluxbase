import js from '@eslint/js'
import tseslint from 'typescript-eslint'

// Flat config for the core TypeScript SDK. Runs type-info-free so it stays
// fast and does not need a tsconfig project service. `bun run lint` lints
// src/ only; framework packages (sdk-react, sdk-vue, sdk-svelte, sdk-next)
// have equivalent configs in their own directories.
//
// The published surface of this package predates lint enforcement, so the
// rules below carry a large backlog of violations. They are deliberately
// `warn` (visible in output, non-gating) instead of `off` so new code is
// nudged away from the pattern. Tightening any of them back to `error`
// requires cleaning up src/ first — do that per rule, not all at once.
export default tseslint.config(
  { ignores: ['dist', 'node_modules', 'coverage'] },
  {
    files: ['src/**/*.ts'],
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    rules: {
      // TODO(lint): ~150 violations across the public API surface (options
      // objects, callbacks). Fix by typing the boundaries properly.
      '@typescript-eslint/no-explicit-any': 'warn',
      // TODO(lint): ~90 violations (unused destructured `data`, unused
      // callback params in fetch wrappers). Prefix with `_` or remove.
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
      // TODO(lint): 1 violation (upload chunking helper).
      '@typescript-eslint/no-this-alias': 'warn',
      // TODO(lint): 4 violations (throw without attaching the caught error
      // as `cause`). New rule introduced with ESLint 10's js.configs.recommended.
      'preserve-caught-error': 'warn',
    },
  },
)
