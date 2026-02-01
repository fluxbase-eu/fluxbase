---
editUrl: false
next: false
prev: false
title: "ChunkedUploadSession"
---

Chunked upload session information

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="bucket"></a> `bucket` | `string` | Target bucket |
| <a id="chunksize"></a> `chunkSize` | `number` | Chunk size used |
| <a id="completedchunks"></a> `completedChunks` | `number`[] | Array of completed chunk indices (0-indexed) |
| <a id="createdat"></a> `createdAt` | `string` | Session creation time |
| <a id="expiresat"></a> `expiresAt` | `string` | Session expiration time |
| <a id="path"></a> `path` | `string` | Target file path |
| <a id="sessionid"></a> `sessionId` | `string` | Unique session identifier for resume |
| <a id="status"></a> `status` | `"active"` \| `"completing"` \| `"completed"` \| `"aborted"` \| `"expired"` | Session status |
| <a id="totalchunks"></a> `totalChunks` | `number` | Total number of chunks |
| <a id="totalsize"></a> `totalSize` | `number` | Total file size |
