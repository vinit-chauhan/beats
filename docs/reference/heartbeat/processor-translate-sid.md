---
navigation_title: "translate_sid"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/heartbeat/current/processor-translate-sid.html
applies_to:
  stack: ga
---

# Translate SID [processor-translate-sid]


The `translate_sid` processor translates a Windows security identifier (SID) into an account name. It retrieves the name of the account associated with the SID, the first domain on which the SID is found, and the type of account. This is only available on Windows.

Every account on a network is issued a unique SID when the account is first created. Internal processes in Windows refer to an account’s SID rather than the account’s user or group name and these values sometimes appear in logs.

If the SID is invalid (malformed) or does not map to any account on the local system or domain then this will result in the processor returning an error unless `ignore_failure` is set.

```yaml
processors:
  - translate_sid:
      field: winlog.event_data.MemberSid
      account_name_target: user.name
      domain_target: user.domain
      ignore_missing: true
      ignore_failure: true
```

The `translate_sid` processor has the following configuration settings:

| Name | Required | Default | Description |
| --- | --- | --- | --- |
| `field` | yes |  | Source field containing a Windows security identifier (SID). |
| `account_name_target` | yes* |  | Target field for the account name value. |
| `account_type_target` | yes* |  | Target field for the account type value. |
| `domain_target` | yes\* |  | Target field for the domain value. |
| `ignore_missing` | no | false | Ignore errors when the source field is missing. |
| `ignore_failure` | no | false | Ignore all errors produced by the processor. |

\* At least one of `account_name_target`, `account_type_target`, and `domain_target` is required to be configured.

