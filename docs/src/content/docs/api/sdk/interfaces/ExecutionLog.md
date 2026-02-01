---
editUrl: false
next: false
prev: false
title: "ExecutionLog"
---

Execution log entry (shared by jobs, RPC, and functions)

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="execution_id"></a> `execution_id` | `string` | ID of the execution (job ID, RPC execution ID, or function execution ID) |
| <a id="fields"></a> `fields?` | `Record`\<`string`, `unknown`\> | Additional structured fields |
| <a id="id"></a> `id` | `number` | Unique log entry ID |
| <a id="level"></a> `level` | `string` | Log level (debug, info, warn, error) |
| <a id="line_number"></a> `line_number` | `number` | Line number within the execution log |
| <a id="message"></a> `message` | `string` | Log message content |
| <a id="timestamp"></a> `timestamp` | `string` | Timestamp of the log entry |
