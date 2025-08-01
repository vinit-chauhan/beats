---
mapped_pages:
  - https://www.elastic.co/guide/en/beats/heartbeat/current/configuration-autodiscover-hints.html
applies_to:
  stack: ga
---

# Hints based autodiscover [configuration-autodiscover-hints]

Heartbeat supports autodiscover based on hints from the provider. The hints system looks for hints in Kubernetes Pod annotations or Docker labels that have the prefix `co.elastic.monitor`. As soon as the container starts, Heartbeat will check if it contains any hints and launch the proper config for it. Hints tell Heartbeat how to get logs for the given container. By default monitors will be created for the container that exposes the port being requested to be monitored. You can use hints to modify this behavior. This is the full list of supported hints:


### `co.elastic.monitor/type` [_co_elastic_monitortype]

Define the monitor type to use. Ex: http, tcp, icmp


### `co.elastic.monitor/hosts` [_co_elastic_monitorhosts]

The URIs to monitor. Example:

```yaml
co.elastic.monitor/type: icmp
co.elastic.monitor/hosts: ${data.host}
```

This would ensure that each host has an ICMP monitor enabled on it.


### `co.elastic.monitor/schedule` [_co_elastic_monitorschedule]

Define the schedule on which the monitor should be executed.

```
co.elastic.monitor/schedule: "@every 5s"
```


### `co.elastic.monitor/processors` [_co_elastic_monitorprocessors]

Define a processor to be added to the Heartbeat monitor configuration. See [Processors](/reference/heartbeat/filtering-enhancing-data.md) for the list of supported processors.

In order to provide ordering of the processor definition, numbers can be provided. If not, the hints builder will do arbitrary ordering:

```yaml
co.elastic.monitor/processors.1.drop_fields.fields: "field1, field2"
co.elastic.monitor/processors.drop_fields.fields: "field3"
```

In the above sample the processor definition tagged with `1` would be executed first.

When hints are used along with templates, then hints will be evaluated only in case there is no template’s condition that resolves to true. For example:

```yaml
heartbeat.autodiscover:
    - type: docker
      hints.enabled: true
      templates:
        - condition:
            contains:
              docker.container.image: redis
          config:
            - type: tcp
              hosts: ["${data.host}:${data.port}"]
              schedule: "@every 1s"
              timeout: 1s
```

In this example first the condition `docker.container.image: redis` is evaluated and if not matched the hints will be processed and if there is again no valid config the `hints.default_config` will be used.


## Kubernetes [_kubernetes_2]

Kubernetes autodiscover provider supports hints in Pod annotations. To enable it just set `hints.enabled`:

```yaml
heartbeat.autodiscover:
  providers:
    - type: kubernetes
      hints.enabled: true
```

You can annotate Kubernetes Pods with useful info to spin up Heartbeat monitors:

```yaml
annotations:
  co.elastic.monitor/type: icmp
  co.elastic.monitor/hosts: ${data.host}
  co.elastic.monitor/schedule: "@every 5s"
```


### Multiple containers [_multiple_containers]

When a pod has multiple containers, the settings are shared unless you add the container name in the hint. For example, these hints configure the container exposing port 8080 to do a HTTP check and have the `sidecar` container to have an TCP check.

```yaml
annotations:
  co.elastic.monitor/type: http
  co.elastic.monitor/hosts: ${data.host}:8080/healthz
  co.elastic.monitor/schedule: "@every 5s"
  co.elastic.monitor.sidecar/type: tcp
  co.elastic.monitor.sidecar/hosts: ${data.host}:8081
  co.elastic.monitor.sidecar/schedule: "@every 5s"
```


### Multiple sets of hints [_multiple_sets_of_hints]

When a container needs multiple monitors to be defined on it, sets of annotations can be provided with numeric prefixes. Annotations without numeric prefixes would default into a single monitor configuration.

```yaml
annotations:
  co.elastic.monitor/type: http
  co.elastic.monitor/hosts: ${data.host}:8080/healthz
  co.elastic.monitor/schedule: "@every 5s"
  co.elastic.monitor/1.type: tcp
  co.elastic.monitor/1.hosts: ${data.host}:8080
  co.elastic.monitor/1.schedule: "@every 5s"
```


## Docker [_docker_3]

Docker autodiscover provider supports hints in labels. To enable it just set `hints.enabled`:

```yaml
heartbeat.autodiscover:
  providers:
    - type: docker
      hints.enabled: true
```

You can label Docker containers with useful info to spin up Heartbeat monitors similar to the Kubernetes example:

```
LABEL co.elastic.monitor/1.type=tcp co.elastic.monitor/1.hosts='${data.host}:6379' co.elastic.monitor/1.schedule='@every 10s'
LABEL co.elastic.monitor/2.type=icmp co.elastic.monitor/2.hosts='${data.host}' co.elastic.monitor/2.schedule='@every 10s'
```

