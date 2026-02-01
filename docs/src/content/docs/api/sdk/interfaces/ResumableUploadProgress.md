---
editUrl: false
next: false
prev: false
title: "ResumableUploadProgress"
---

Upload progress information for resumable uploads

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="bytespersecond"></a> `bytesPerSecond` | `number` | Transfer rate in bytes per second |
| <a id="currentchunk"></a> `currentChunk` | `number` | Current chunk being uploaded (1-indexed) |
| <a id="loaded"></a> `loaded` | `number` | Number of bytes uploaded so far |
| <a id="percentage"></a> `percentage` | `number` | Upload percentage (0-100) |
| <a id="sessionid"></a> `sessionId` | `string` | Upload session ID (for resume capability) |
| <a id="total"></a> `total` | `number` | Total file size in bytes |
| <a id="totalchunks"></a> `totalChunks` | `number` | Total number of chunks |
