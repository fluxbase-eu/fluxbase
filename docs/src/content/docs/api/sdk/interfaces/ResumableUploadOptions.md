---
editUrl: false
next: false
prev: false
title: "ResumableUploadOptions"
---

Options for resumable chunked uploads

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="cachecontrol"></a> `cacheControl?` | `string` | Cache-Control header value |
| <a id="chunksize"></a> `chunkSize?` | `number` | Chunk size in bytes for each upload request. **Default** `5242880 (5MB)` |
| <a id="chunktimeout"></a> `chunkTimeout?` | `number` | Timeout in milliseconds per chunk request. **Default** `60000 (1 minute)` |
| <a id="contenttype"></a> `contentType?` | `string` | MIME type of the file |
| <a id="maxretries"></a> `maxRetries?` | `number` | Number of retry attempts per chunk on failure. **Default** `3` |
| <a id="metadata"></a> `metadata?` | `Record`\<`string`, `string`\> | Custom metadata to attach to the file |
| <a id="onprogress"></a> `onProgress?` | (`progress`) => `void` | Callback for upload progress |
| <a id="resumesessionid"></a> `resumeSessionId?` | `string` | Existing upload session ID to resume (optional) |
| <a id="retrydelayms"></a> `retryDelayMs?` | `number` | Base delay in milliseconds for exponential backoff. **Default** `1000` |
| <a id="signal"></a> `signal?` | `AbortSignal` | AbortSignal to cancel the upload |
