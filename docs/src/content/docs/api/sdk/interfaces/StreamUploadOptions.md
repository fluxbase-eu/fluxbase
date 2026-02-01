---
editUrl: false
next: false
prev: false
title: "StreamUploadOptions"
---

Options for streaming uploads (memory-efficient for large files)

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="cachecontrol"></a> `cacheControl?` | `string` | Cache-Control header value |
| <a id="contenttype"></a> `contentType?` | `string` | MIME type of the file |
| <a id="metadata"></a> `metadata?` | `Record`\<`string`, `string`\> | Custom metadata to attach to the file |
| <a id="onuploadprogress"></a> `onUploadProgress?` | (`progress`) => `void` | Optional callback to track upload progress |
| <a id="signal"></a> `signal?` | `AbortSignal` | AbortSignal to cancel the upload |
| <a id="upsert"></a> `upsert?` | `boolean` | If true, overwrite existing file at this path |
