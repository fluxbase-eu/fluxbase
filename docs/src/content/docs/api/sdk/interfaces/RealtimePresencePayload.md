---
editUrl: false
next: false
prev: false
title: "RealtimePresencePayload"
---

Realtime presence payload structure

## Properties

| Property | Type |
| ------ | ------ |
| <a id="currentpresences"></a> `currentPresences?` | `Record`\<`string`, [`PresenceState`](/api/sdk/interfaces/presencestate/)[]\> |
| <a id="event"></a> `event` | `"sync"` \| `"join"` \| `"leave"` |
| <a id="key"></a> `key?` | `string` |
| <a id="leftpresences"></a> `leftPresences?` | [`PresenceState`](/api/sdk/interfaces/presencestate/)[] |
| <a id="newpresences"></a> `newPresences?` | [`PresenceState`](/api/sdk/interfaces/presencestate/)[] |
