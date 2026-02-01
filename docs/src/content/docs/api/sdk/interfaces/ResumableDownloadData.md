---
editUrl: false
next: false
prev: false
title: "ResumableDownloadData"
---

Response type for resumable downloads - stream abstracts chunking

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="size"></a> `size` | `number` \| `null` | File size in bytes from HEAD request, or null if unknown |
| <a id="stream"></a> `stream` | `ReadableStream`\<`Uint8Array`\<`ArrayBufferLike`\>\> | The readable stream for the file content (abstracts chunking internally) |
