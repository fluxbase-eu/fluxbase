---
editUrl: false
next: false
prev: false
title: "ResumableDownloadOptions"
---

Options for resumable chunked downloads

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="chunksize"></a> `chunkSize?` | `number` | Chunk size in bytes for each download request. **Default** `5242880 (5MB)` |
| <a id="chunktimeout"></a> `chunkTimeout?` | `number` | Timeout in milliseconds per chunk request. **Default** `30000` |
| <a id="maxretries"></a> `maxRetries?` | `number` | Number of retry attempts per chunk on failure. **Default** `3` |
| <a id="onprogress"></a> `onProgress?` | (`progress`) => `void` | Callback for download progress |
| <a id="retrydelayms"></a> `retryDelayMs?` | `number` | Base delay in milliseconds for exponential backoff. **Default** `1000` |
| <a id="signal"></a> `signal?` | `AbortSignal` | AbortSignal to cancel the download |
