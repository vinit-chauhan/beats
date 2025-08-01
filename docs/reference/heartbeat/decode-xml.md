---
navigation_title: "decode_xml"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/heartbeat/current/decode-xml.html
applies_to:
  stack: ga
---

# Decode XML [decode-xml]


The `decode_xml` processor decodes XML data that is stored under the `field` key. It outputs the result into the `target_field`.

This example demonstrates how to decode an XML string contained in the `message` field and write the resulting fields into the root of the document. Any fields that already exist will be overwritten.

```yaml
processors:
  - decode_xml:
      field: message
      target_field: ""
      overwrite_keys: true
```

By default any decoding errors that occur will stop the processing chain and the error will be added to `error.message` field. To ignore all errors and continue to the next processor you can set `ignore_failure: true`. To specifically ignore failures caused by `field` not existing you can set `ignore_missing: true`.

```yaml
processors:
  - decode_xml:
      field: example
      target_field: xml
      ignore_missing: true
      ignore_failure: true
```

By default all keys converted from XML will have the names converted to lowercase. If there is a need to disable this behavior it is possible to use the below example:

```yaml
processors:
  - decode_xml:
      field: message
      target_field: xml
      to_lower: false
```

Example XML input:

```xml
<catalog>
  <book seq="1">
    <author>William H. Gaddis</author>
    <title>The Recognitions</title>
    <review>One of the great seminal American novels of the 20th century.</review>
  </book>
</catalog>
```

Will produce the following output:

```json
{
	"xml": {
		"catalog": {
			"book": {
				"author": "William H. Gaddis",
				"review": "One of the great seminal American novels of the 20th century.",
				"seq": "1",
				"title": "The Recognitions"
			}
		}
	}
}
```

The supported configuration options are:

`field`
:   (Required) Source field containing the XML. Defaults to `message`.

`target_field`
:   (Optional) The field under which the decoded XML will be written. By default the decoded XML object replaces the field from which it was read. To merge the decoded XML fields into the root of the event specify `target_field` with an empty string (`target_field: ""`). Note that the `null` value (`target_field:`) is treated as if the field was not set at all.

`overwrite_keys`
:   (Optional) A boolean that specifies whether keys that already exist in the event are overwritten by keys from the decoded XML object. The default value is `true`.

`to_lower`
:   (Optional) Converts all keys to lowercase. Accepts either `true` or `false`. The default value is `true`.

`document_id`
:   (Optional) XML key to use as the document ID. If configured, the field will be removed from the original XML document and stored in `@metadata._id`.

`ignore_missing`
:   (Optional) If `true` the processor will not return an error when a specified field does not exist. Defaults to `false`.

`ignore_failure`
:   (Optional) Ignore all errors produced by the processor. Defaults to `false`.

See [Conditions](/reference/heartbeat/defining-processors.md#conditions) for a list of supported conditions.

