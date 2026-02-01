---
editUrl: false
next: false
prev: false
title: "DownloadOptions"
---

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="signal"></a> `signal?` | `AbortSignal` | AbortSignal to cancel the download |
| <a id="stream"></a> `stream?` | `boolean` | If true, returns a ReadableStream instead of Blob |
| <a id="timeout"></a> `timeout?` | `number` | Timeout in milliseconds for the download request. For streaming downloads, this applies to the initial response. Set to 0 or undefined for no timeout (recommended for large files). **Default** `undefined (no timeout for streaming, 30000 for non-streaming)` |
