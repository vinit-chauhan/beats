---
navigation_title: "move_fields"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/metricbeat/current/move-fields.html
applies_to:
  stack: ga
---

# Move fields [move-fields]


The `move_fields` processor moves event fields from one object into another. It can also rearrange fields or add a prefix to fields.

The processor extracts fields from `from`, then uses `fields` and `exclude` as filters to choose which fields to move into the `to` field.

For example, given the following event:

```json
{
  "app": {
    "method": "a",
    "elapsed_time": 100,
    "user_id": 100,
    "message": "i'm a message"
  }
}
```

To move `method` and `elapsed_time` into another object, use this configuration:

```yaml
processors:
  - move_fields:
      from: "app"
      fields: ["method", "elapsed_time"],
      to: "rpc."
```

Your final event will be:

```json
{
  "app": {
    "user_id": 100,
    "message": "i'm a message",
    "rpc": {
      "method": "a",
      "elapsed_time": 100
    }
  }
}
```

To add a prefix to the whole event:

```json
{
  "app": { "method": "a"},
  "cost": 100
}
```

Use this configuration:

```yaml
processors:
  - move_fields:
      to: "my_prefix_"
```

Your final event will be:

```json
{
  "my_prefix_app": { "method": "a"},
  "my_prefix_cost": 100
}
```

| Name | Required | Default | Description |
| --- | --- | --- | --- |
| `from` | no |  | Which field you want extract. This field and any nested fields will be moved into `to` unless they are filtered out. If empty, indicates event root. |
| `fields` | no |  | Which fields to extract from `from` and move to `to`. An empty list indicates all fields. |
| `ignore_missing` | no | false | Ignore "not found" errors when extracting fields. |
| `exclude` | no |  | A list of fields to exclude and not move. |
| `to` | yes |  | These fields extract from `from` destination field prefix the `to` will base on fields root. |

```yaml
processors:
  - move_fields:
      from: "app"
      fields: [ "method", "elapsed_time" ]
      to: "rpc."
```

