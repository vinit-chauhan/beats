---
navigation_title: "decode_xml_wineventlog"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/packetbeat/current/decode-xml-wineventlog.html
applies_to:
  stack: preview
---

# Decode XML Wineventlog [decode-xml-wineventlog]


::::{warning}
This functionality is in technical preview and may be changed or removed in a future release. Elastic will work to fix any issues, but features in technical preview are not subject to the support SLA of official GA features.
::::


The `decode_xml_wineventlog` processor decodes Windows Event Log data in XML format that is stored under the `field` key. It outputs the result into the `target_field`.

The output fields will be the same as the [winlogbeat winlog fields](/reference/winlogbeat/exported-fields-winlog.md#_winlog).

The supported configuration options are:

`field`
:   (Required) Source field containing the XML. Defaults to `message`.

`target_field`
:   (Required) The field under which the decoded XML will be written. To merge the decoded XML fields into the root of the event specify `target_field` with an empty string (`target_field: ""`). The default value is `winlog`.

`overwrite_keys`
:   (Optional) A boolean that specifies whether keys that already exist in the event are overwritten by keys from the decoded XML object. The default value is `true`.

`map_ecs_fields`
:   (Optional) A boolean that specifies whether to map additional ECS fields when possible. Note that ECS field keys are placed outside of `target_field`. The default value is `true`.

`ignore_missing`
:   (Optional) If `true` the processor will not return an error when a specified field does not exist. Defaults to `false`.

`ignore_failure`
:   (Optional) Ignore all errors produced by the processor. Defaults to `false`.

`language`
:   (Optional) The language ID the events will be rendered in. The language will be forced regardless of the system language. Forwarded events will ignore this setting. A complete list of language IDs can be found [here](https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-lcid/a9eac961-e77d-41a6-90a5-ce1a8b0cdb9c). It defaults to `0`, which indicates to use the system language.

Example:

```yaml
processors:
  - decode_xml_wineventlog:
      field: event.original
      target_field: winlog
```

```json
{
  "event": {
    "original": "<Event xmlns='http://schemas.microsoft.com/win/2004/08/events/event'><System><Provider Name='Microsoft-Windows-Security-Auditing' Guid='{54849625-5478-4994-a5ba-3e3b0328c30d}'/><EventID>4672</EventID><Version>0</Version><Level>0</Level><Task>12548</Task><Opcode>0</Opcode><Keywords>0x8020000000000000</Keywords><TimeCreated SystemTime='2021-03-23T09:56:13.137310000Z'/><EventRecordID>11303</EventRecordID><Correlation ActivityID='{ffb23523-1f32-0000-c335-b2ff321fd701}'/><Execution ProcessID='652' ThreadID='4660'/><Channel>Security</Channel><Computer>vagrant</Computer><Security/></System><EventData><Data Name='SubjectUserSid'>S-1-5-18</Data><Data Name='SubjectUserName'>SYSTEM</Data><Data Name='SubjectDomainName'>NT AUTHORITY</Data><Data Name='SubjectLogonId'>0x3e7</Data><Data Name='PrivilegeList'>SeAssignPrimaryTokenPrivilege\n\t\t\tSeTcbPrivilege\n\t\t\tSeSecurityPrivilege\n\t\t\tSeTakeOwnershipPrivilege\n\t\t\tSeLoadDriverPrivilege\n\t\t\tSeBackupPrivilege\n\t\t\tSeRestorePrivilege\n\t\t\tSeDebugPrivilege\n\t\t\tSeAuditPrivilege\n\t\t\tSeSystemEnvironmentPrivilege\n\t\t\tSeImpersonatePrivilege\n\t\t\tSeDelegateSessionUserImpersonatePrivilege</Data></EventData><RenderingInfo Culture='en-US'><Message>Special privileges assigned to new logon.\n\nSubject:\n\tSecurity ID:\t\tS-1-5-18\n\tAccount Name:\t\tSYSTEM\n\tAccount Domain:\t\tNT AUTHORITY\n\tLogon ID:\t\t0x3E7\n\nPrivileges:\t\tSeAssignPrimaryTokenPrivilege\n\t\t\tSeTcbPrivilege\n\t\t\tSeSecurityPrivilege\n\t\t\tSeTakeOwnershipPrivilege\n\t\t\tSeLoadDriverPrivilege\n\t\t\tSeBackupPrivilege\n\t\t\tSeRestorePrivilege\n\t\t\tSeDebugPrivilege\n\t\t\tSeAuditPrivilege\n\t\t\tSeSystemEnvironmentPrivilege\n\t\t\tSeImpersonatePrivilege\n\t\t\tSeDelegateSessionUserImpersonatePrivilege</Message><Level>Information</Level><Task>Special Logon</Task><Opcode>Info</Opcode><Channel>Security</Channel><Provider>Microsoft Windows security auditing.</Provider><Keywords><Keyword>Audit Success</Keyword></Keywords></RenderingInfo></Event>"
  }
}
```

Will produce the following output:

```json
{
  "event": {
    "original": "<Event xmlns='http://schemas.microsoft.com/win/2004/08/events/event'><System><Provider Name='Microsoft-Windows-Security-Auditing' Guid='{54849625-5478-4994-a5ba-3e3b0328c30d}'/><EventID>4672</EventID><Version>0</Version><Level>0</Level><Task>12548</Task><Opcode>0</Opcode><Keywords>0x8020000000000000</Keywords><TimeCreated SystemTime='2021-03-23T09:56:13.137310000Z'/><EventRecordID>11303</EventRecordID><Correlation ActivityID='{ffb23523-1f32-0000-c335-b2ff321fd701}'/><Execution ProcessID='652' ThreadID='4660'/><Channel>Security</Channel><Computer>vagrant</Computer><Security/></System><EventData><Data Name='SubjectUserSid'>S-1-5-18</Data><Data Name='SubjectUserName'>SYSTEM</Data><Data Name='SubjectDomainName'>NT AUTHORITY</Data><Data Name='SubjectLogonId'>0x3e7</Data><Data Name='PrivilegeList'>SeAssignPrimaryTokenPrivilege\n\t\t\tSeTcbPrivilege\n\t\t\tSeSecurityPrivilege\n\t\t\tSeTakeOwnershipPrivilege\n\t\t\tSeLoadDriverPrivilege\n\t\t\tSeBackupPrivilege\n\t\t\tSeRestorePrivilege\n\t\t\tSeDebugPrivilege\n\t\t\tSeAuditPrivilege\n\t\t\tSeSystemEnvironmentPrivilege\n\t\t\tSeImpersonatePrivilege\n\t\t\tSeDelegateSessionUserImpersonatePrivilege</Data></EventData><RenderingInfo Culture='en-US'><Message>Special privileges assigned to new logon.\n\nSubject:\n\tSecurity ID:\t\tS-1-5-18\n\tAccount Name:\t\tSYSTEM\n\tAccount Domain:\t\tNT AUTHORITY\n\tLogon ID:\t\t0x3E7\n\nPrivileges:\t\tSeAssignPrimaryTokenPrivilege\n\t\t\tSeTcbPrivilege\n\t\t\tSeSecurityPrivilege\n\t\t\tSeTakeOwnershipPrivilege\n\t\t\tSeLoadDriverPrivilege\n\t\t\tSeBackupPrivilege\n\t\t\tSeRestorePrivilege\n\t\t\tSeDebugPrivilege\n\t\t\tSeAuditPrivilege\n\t\t\tSeSystemEnvironmentPrivilege\n\t\t\tSeImpersonatePrivilege\n\t\t\tSeDelegateSessionUserImpersonatePrivilege</Message><Level>Information</Level><Task>Special Logon</Task><Opcode>Info</Opcode><Channel>Security</Channel><Provider>Microsoft Windows security auditing.</Provider><Keywords><Keyword>Audit Success</Keyword></Keywords></RenderingInfo></Event>",
    "action":   "Special Logon",
		"code":     "4672",
		"kind":     "event",
		"outcome":  "success",
		"provider": "Microsoft-Windows-Security-Auditing",
  },
	"host": {
    "name": "vagrant",
  },
  "log": {
    "level": "information",
  },
  "winlog": {
    "channel": "Security",
    "outcome": "success",
    "activity_id": "{ffb23523-1f32-0000-c335-b2ff321fd701}",
    "level": "information",
    "event_id": 4672,
    "provider_name": "Microsoft-Windows-Security-Auditing",
    "record_id": 11303,
    "computer_name": "vagrant",
    "keywords_raw": 9232379236109516800,
    "opcode": "Info",
    "provider_guid": "{54849625-5478-4994-a5ba-3e3b0328c30d}",
    "event_data": {
      "SubjectUserSid": "S-1-5-18",
      "SubjectUserName": "SYSTEM",
      "SubjectDomainName": "NT AUTHORITY",
      "SubjectLogonId": "0x3e7",
      "PrivilegeList": "SeAssignPrimaryTokenPrivilege\n\t\t\tSeTcbPrivilege\n\t\t\tSeSecurityPrivilege\n\t\t\tSeTakeOwnershipPrivilege\n\t\t\tSeLoadDriverPrivilege\n\t\t\tSeBackupPrivilege\n\t\t\tSeRestorePrivilege\n\t\t\tSeDebugPrivilege\n\t\t\tSeAuditPrivilege\n\t\t\tSeSystemEnvironmentPrivilege\n\t\t\tSeImpersonatePrivilege\n\t\t\tSeDelegateSessionUserImpersonatePrivilege"
    },
    "task": "Special Logon",
    "keywords": [
      "Audit Success"
    ],
    "message": "Special privileges assigned to new logon.\n\nSubject:\n\tSecurity ID:\t\tS-1-5-18\n\tAccount Name:\t\tSYSTEM\n\tAccount Domain:\t\tNT AUTHORITY\n\tLogon ID:\t\t0x3E7\n\nPrivileges:\t\tSeAssignPrimaryTokenPrivilege\n\t\t\tSeTcbPrivilege\n\t\t\tSeSecurityPrivilege\n\t\t\tSeTakeOwnershipPrivilege\n\t\t\tSeLoadDriverPrivilege\n\t\t\tSeBackupPrivilege\n\t\t\tSeRestorePrivilege\n\t\t\tSeDebugPrivilege\n\t\t\tSeAuditPrivilege\n\t\t\tSeSystemEnvironmentPrivilege\n\t\t\tSeImpersonatePrivilege\n\t\t\tSeDelegateSessionUserImpersonatePrivilege",
    "process": {
      "pid": 652,
      "thread": {
        "id": 4660
      }
    }
  }
}
```

See [Conditions](/reference/packetbeat/defining-processors.md#conditions) for a list of supported conditions.

The field mappings are as follows:

| Event Field | Source XML Element | Notes |
| --- | --- | --- |
| `winlog.channel` | `<Event><System><Channel>` |  |
| `winlog.event_id` | `<Event><System><EventID>` |  |
| `winlog.provider_name` | `<Event><System><Provider>` | `Name` attribute |
| `winlog.record_id` | `<Event><System><EventRecordID>` |  |
| `winlog.task` | `<Event><System><Task>` |  |
| `winlog.computer_name` | `<Event><System><Computer>` |  |
| `winlog.keywords` | `<Event><RenderingInfo><Keywords>` | list of each `Keyword` |
| `winlog.opcodes` | `<Event><RenderingInfo><Opcode>` |  |
| `winlog.provider_guid` | `<Event><System><Provider>` | `Guid` attribute |
| `winlog.version` | `<Event><System><Version>` |  |
| `winlog.time_created` | `<Event><System><TimeCreated>` | `SystemTime` attribute |
| `winlog.outcome` | `<Event><System><Keywords>` | "success" if bit 0x20000000000000 is set, "failure" if 0x10000000000000 is set |
| `winlog.level` | `<Event><System><Level>` | converted to lowercase |
| `winlog.message` | `<Event><RenderingInfo><Message>` | line endings removed |
| `winlog.user.identifier` | `<Event><System><Security><UserID>` |  |
| `winlog.user.domain` | `<Event><System><Security><Domain>` |  |
| `winlog.user.name` | `<Event><System><Security><Name>` |  |
| `winlog.user.type` | `<Event><System><Security><Type>` | converted from integer to String |
| `winlog.event_data` | `<Event><EventData>` | map where `Name` attribute in Data element is key, and value is the value of the Data element |
| `winlog.user_data` | `<Event><UserData>` | map where `Name` attribute in Data element is key, and value is the value of the Data element |
| `winlog.activity_id` | `<Event><System><Correlation><ActivityID>` |  |
| `winlog.related_activity_id` | `<Event><System><Correlation><RelatedActivityID>` |  |
| `winlog.kernel_time` | `<Event><System><Execution><KernelTime>` |  |
| `winlog.process.pid` | `<Event><System><Execution><ProcessID>` |  |
| `winlog.process.thread.id` | `<Event><System><Execution><ThreadID>` |  |
| `winlog.processor_id` | `<Event><System><Execution><ProcessorID>` |  |
| `winlog.processor_time` | `<Event><System><Execution><ProcessorTime>` |  |
| `winlog.session_id` | `<Event><System><Execution><SessionID>` |  |
| `winlog.user_time` | `<Event><System><Execution><UserTime>` |  |
| `winlog.error.code` | `<Event><ProcessingErrorData><ErrorCode>` |  |

If `map_ecs_fields` is enabled then the following field mappings are also performed:

| Event Field | Source XML or other field | Notes |
| --- | --- | --- |
| `event.code` | `winlog.event_id` |  |
| `event.kind` | `"event"` |  |
| `event.provider` | `<Event><System><Provider>` | `Name` attribute |
| `event.action` | `<Event><RenderingInfo><Task>` |  |
| `event.host.name` | `<Event><System><Computer>` |  |
| `event.outcome` | `winlog.outcome` |  |
| `log.level` | `winlog.level` |  |
| `message` | `winlog.message` |  |
| `error.code` | `winlog.error.code` |  |
| `error.message` | `winlog.error.message` |  |

