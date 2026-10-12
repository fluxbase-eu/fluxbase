---
editUrl: false
next: false
prev: false
title: "UploadOptions"
---

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="cachecontrol"></a> ~~`cacheControl?`~~ | `string` | :::caution[Deprecated] Not read by the server's upload handler; accepted for backwards compatibility but not transmitted. ::: |
| <a id="contenttype"></a> `contentType?` | `string` | - |
| <a id="metadata"></a> `metadata?` | `Record`\<`string`, `string`\> | Custom metadata to attach to the file. Sent as `metadata_<key>` form fields (the only form of metadata the server's upload handler reads). |
| <a id="onuploadprogress"></a> `onUploadProgress?` | (`progress`) => `void` | Optional callback to track upload progress |
| <a id="upsert"></a> ~~`upsert?`~~ | `boolean` | :::caution[Deprecated] Not read by the server's upload handler; uploads always upsert (ON CONFLICT (bucket_id, path) DO UPDATE). ::: |
