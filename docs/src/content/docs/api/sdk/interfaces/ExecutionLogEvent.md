---
editUrl: false
next: false
prev: false
title: "ExecutionLogEvent"
---

Execution log event received from realtime subscription

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="execution_id"></a> `execution_id` | `string` | Unique execution ID |
| <a id="execution_type"></a> `execution_type` | [`ExecutionType`](/api/sdk/type-aliases/executiontype/) | Type of execution |
| <a id="fields"></a> `fields?` | `Record`\<`string`, `unknown`\> | Additional fields |
| <a id="level"></a> `level` | [`ExecutionLogLevel`](/api/sdk/type-aliases/executionloglevel/) | Log level |
| <a id="line_number"></a> `line_number` | `number` | Line number in the execution log |
| <a id="message"></a> `message` | `string` | Log message content |
| <a id="timestamp"></a> `timestamp` | `string` | Timestamp of the log entry |
