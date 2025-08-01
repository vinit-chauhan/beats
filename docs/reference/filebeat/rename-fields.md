---
navigation_title: "rename"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/filebeat/current/rename-fields.html
applies_to:
  stack: ga
---

# Rename fields from events [rename-fields]


The `rename` processor specifies a list of fields to rename. Under the `fields` key, each entry contains a `from: old-key` and a `to: new-key` pair, where:

* `from` is the original field name. It’s supported to use `@metadata.` prefix for `from` and rename keys in the event metadata instead of event fields.
* `to` is the target field name

The `rename` processor cannot be used to overwrite fields. To overwrite fields either first rename the target field, or use the `drop_fields` processor to drop the field and then rename the field.

::::{tip}
You can rename fields to resolve field name conflicts. For example, if an event has two fields, `c` and `c.b` (where `b` is a subfield of `c`), assigning scalar values results in an {{es}} error at ingest time. The assignment `{"c": 1, "c.b": 2}` would result in an error because `c` is an object and cannot be assigned a scalar value. To prevent this conflict, rename `c` to `c.value` before assigning values.
::::


```yaml
processors:
  - rename:
      fields:
        - from: "a.g"
          to: "e.d"
      ignore_missing: false
      fail_on_error: true
```

The `rename` processor has the following configuration settings:

`ignore_missing`
:   (Optional) If set to true, no error is logged in case a key which should be renamed is missing. Default is `false`.

`fail_on_error`
:   (Optional) If set to true, in case of an error the renaming of fields is stopped and the original event is returned. If set to false, renaming continues also if an error happened during renaming. Default is `true`.

See [Conditions](/reference/filebeat/defining-processors.md#conditions) for a list of supported conditions.

You can specify multiple `rename` processors under the `processors` section.

