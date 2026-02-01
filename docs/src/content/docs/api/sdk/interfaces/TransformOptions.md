---
editUrl: false
next: false
prev: false
title: "TransformOptions"
---

Options for on-the-fly image transformations
Applied to storage downloads via query parameters

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="fit"></a> `fit?` | [`ImageFitMode`](/api/sdk/type-aliases/imagefitmode/) | How to fit the image within target dimensions (default: cover) |
| <a id="format"></a> `format?` | [`ImageFormat`](/api/sdk/type-aliases/imageformat/) | Output format (defaults to original format) |
| <a id="height"></a> `height?` | `number` | Target height in pixels (0 or undefined = auto based on width) |
| <a id="quality"></a> `quality?` | `number` | Output quality 1-100 (default: 80) |
| <a id="width"></a> `width?` | `number` | Target width in pixels (0 or undefined = auto based on height) |
