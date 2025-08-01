---
navigation_title: "decode_base64_field"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/heartbeat/current/decode-base64-field.html
applies_to:
  stack: ga
---

# Decode Base64 fields [decode-base64-field]


The `decode_base64_field` processor specifies a field to base64 decode. The `field` key contains a `from: old-key` and a `to: new-key` pair. `from` is the origin and `to` the target name of the field.

To overwrite fields either first rename the target field or use the `drop_fields` processor to drop the field and then rename the field.

```yaml
processors:
  - decode_base64_field:
      field:
        from: "field1"
        to: "field2"
      ignore_missing: false
      fail_on_error: true
```

In the example above:
- field1 is decoded in field2

The `decode_base64_field` processor has the following configuration settings:

`ignore_missing`
:   (Optional) If set to true, no error is logged in case a key which should be base64 decoded is missing. Default is `false`.

`fail_on_error`
:   (Optional) If set to true, in case of an error the base64 decode of fields is stopped and the original event is returned. If set to false, decoding continues also if an error happened during decoding. Default is `true`.

See [Conditions](/reference/heartbeat/defining-processors.md#conditions) for a list of supported conditions.

