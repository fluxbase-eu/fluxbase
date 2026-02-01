---
editUrl: false
next: false
prev: false
title: "CreateBranchOptions"
---

Options for creating a new branch

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="dataclonemode"></a> `dataCloneMode?` | [`DataCloneMode`](/api/sdk/type-aliases/dataclonemode/) | How to clone data |
| <a id="expiresin"></a> `expiresIn?` | `string` | Duration until branch expires (e.g., "24h", "7d") |
| <a id="githubprnumber"></a> `githubPRNumber?` | `number` | GitHub PR number (for preview branches) |
| <a id="githubprurl"></a> `githubPRUrl?` | `string` | GitHub PR URL |
| <a id="githubrepo"></a> `githubRepo?` | `string` | GitHub repository (owner/repo) |
| <a id="parentbranchid"></a> `parentBranchId?` | `string` | Parent branch to clone from (defaults to main) |
| <a id="type"></a> `type?` | [`BranchType`](/api/sdk/type-aliases/branchtype/) | Branch type |
