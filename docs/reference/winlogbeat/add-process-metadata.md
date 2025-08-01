---
navigation_title: "add_process_metadata"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/winlogbeat/current/add-process-metadata.html
applies_to:
  stack: ga
---

# Add process metadata [add-process-metadata]


The `add_process_metadata` processor enriches events with information from running processes, identified by their process ID (PID).

```yaml
processors:
  - add_process_metadata:
      match_pids:
        - process.pid
```

The fields added to the event look as follows:

```json
{
  "container": {
    "id": "b5285682fba7449c86452b89a800609440ecc88a7ba5f2d38bedfb85409b30b1"
  },
  "process": {
    "args": [
      "/usr/lib/systemd/systemd",
      "--switched-root",
      "--system",
      "--deserialize",
      "22"
    ],
    "executable": "/usr/lib/systemd/systemd",
    "name": "systemd",
    "owner": {
      "id": "0",
      "name": "root"
    },
    "parent": {
      "pid": 0
    },
    "pid": 1,
    "start_time": "2018-08-22T08:44:50.684Z",
    "title": "/usr/lib/systemd/systemd --switched-root --system --deserialize 22"
  }
}
```

Optionally, the process environment can be included, too:

```json
  ...
  "env": {
    "HOME":       "/",
    "TERM":       "linux",
    "BOOT_IMAGE": "/boot/vmlinuz-4.11.8-300.fc26.x86_64",
    "LANG":       "en_US.UTF-8",
  }
  ...
```

It has the following settings:

`match_pids`
:   List of fields to lookup for a PID. The processor will search the list sequentially until the field is found in the current event, and the PID lookup will be applied to the value of this field.

`target`
:   (Optional) Destination prefix where the `process` object will be created. The default is the event’s root.

`include_fields`
:   (Optional) List of fields to add. By default, the processor will add all the available fields except `process.env`.

`ignore_missing`
:   (Optional) When set to `false`, events that don’t contain any of the fields in match_pids will be discarded and an error will be generated. By default, this condition is ignored.

`overwrite_keys`
:   (Optional) By default, if a target field already exists, it will not be overwritten, and an error will be logged. If `overwrite_keys` is set to `true`, this condition will be ignored.

`restricted_fields`
:   (Optional) By default, the `process.env` field is not output, to avoid leaking sensitive data. If `restricted_fields` is `true`, the field will be present in the output.

`host_path`
:   (Optional) By default, the `host_path` field is set to the root directory of the host `/`. This is the path where `/proc` is mounted. For different runtime configurations of Kubernetes or Docker, the `host_path` can be set to overwrite the default.

`cgroup_prefixes`
:   (Optional) List of prefixes that will be matched against cgroup paths. When a cgroup path begins with a prefix in the list, then the last element of the path is returned as the container ID. Only one of `cgroup_prefixes` and `cgroup_rexex` should be configured. If neither are configured then a default `cgroup_regex` value is used that matches cgroup paths containing 64-character container IDs (like those from Docker, Kubernetes, and Podman).

`cgroup_regex`
:   (Optional) A regular expression that will be matched against cgroup paths. It must contain one capturing group. When a cgroup path matches the regular expression then the value of the capturing group is returned as the container ID.  Only one of `cgroup_prefixes` and `cgroup_rexex` should be configured. If neither are configured then a default `cgroup_regex` value is used that matches cgroup paths containing 64-character container IDs (like those from Docker, Kubernetes, and Podman).

`cgroup_cache_expire_time`
:   (Optional) By default, the `cgroup_cache_expire_time` is set to 30 seconds. This is the length of time before cgroup cache elements expire in seconds. It can be set to 0 to disable the cgroup cache. In some container runtimes technology like runc, the container’s process is also process in the host kernel, and will be affected by PID rollover/reuse. The expire time needs to set smaller than the PIDs wrap around time to avoid wrong container id.

