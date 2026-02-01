---
editUrl: false
next: false
prev: false
title: "StreamDownloadData"
---

Response type for stream downloads, includes file size from Content-Length header

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="size"></a> `size` | `number` \| `null` | File size in bytes from Content-Length header, or null if unknown |
| <a id="stream"></a> `stream` | `ReadableStream`\<`Uint8Array`\<`ArrayBufferLike`\>\> | The readable stream for the file content |
