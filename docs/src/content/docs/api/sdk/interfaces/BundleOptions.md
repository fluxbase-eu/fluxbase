---
editUrl: false
next: false
prev: false
title: "BundleOptions"
---

Options for bundling code

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="basedir"></a> `baseDir?` | `string` | Base directory for resolving relative imports (resolveDir in esbuild) |
| <a id="code"></a> `code` | `string` | Entry point code |
| <a id="define"></a> `define?` | `Record`\<`string`, `string`\> | Custom define values for esbuild (e.g., { 'process.env.NODE_ENV': '"production"' }) |
| <a id="external"></a> `external?` | `string`[] | External modules to exclude from bundle |
| <a id="importmap"></a> `importMap?` | `Record`\<`string`, `string`\> | Import map from deno.json (maps aliases to npm: or file paths) |
| <a id="minify"></a> `minify?` | `boolean` | Minify output |
| <a id="nodepaths"></a> `nodePaths?` | `string`[] | Additional paths to search for node_modules (useful when importing from parent directories) |
| <a id="sourcemap"></a> `sourcemap?` | `boolean` | Source map generation |
