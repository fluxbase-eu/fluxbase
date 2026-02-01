---
editUrl: false
next: false
prev: false
title: "DownloadProgress"
---

Download progress information

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="bytespersecond"></a> `bytesPerSecond` | `number` | Transfer rate in bytes per second |
| <a id="currentchunk"></a> `currentChunk` | `number` | Current chunk being downloaded (1-indexed) |
| <a id="loaded"></a> `loaded` | `number` | Number of bytes downloaded so far |
| <a id="percentage"></a> `percentage` | `number` \| `null` | Download percentage (0-100), or null if total is unknown |
| <a id="total"></a> `total` | `number` \| `null` | Total file size in bytes, or null if unknown |
| <a id="totalchunks"></a> `totalChunks` | `number` \| `null` | Total number of chunks, or null if total size unknown |
