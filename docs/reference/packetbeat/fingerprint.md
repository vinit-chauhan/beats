---
navigation_title: "fingerprint"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/packetbeat/current/fingerprint.html
applies_to:
  stack: ga
---

# Generate a fingerprint of an event [fingerprint]


The `fingerprint` processor generates a fingerprint of an event based on a specified subset of its fields.

The value that is hashed is constructed as a concatenation of the field name and field value separated by `|`. For example `|field1|value1|field2|value2|`.

Nested fields are supported in the following format: `"field1.field2"` e.g: `["log.path.file", "foo"]`

```yaml
processors:
  - fingerprint:
      fields: ["field1", "field2", ...]
```

The following settings are supported:

`fields`
:   List of fields to use as the source for the fingerprint. The list will be alphabetically sorted by the processor.

`ignore_missing`
:   (Optional) Whether to ignore missing fields. Default is `false`.

`target_field`
:   (Optional) Field in which the generated fingerprint should be stored. Default is `fingerprint`.

`method`
:   (Optional) Algorithm to use for computing the fingerprint. Must be one of: `md5`, `sha1`, `sha256`, `sha384`, `sha512`, `xxhash`. Default is `sha256`.

`encoding`
:   (Optional) Encoding to use on the fingerprint value. Must be one of `hex`, `base32`, or `base64`. Default is `hex`.

