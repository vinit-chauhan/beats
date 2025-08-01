---
navigation_title: "HTTP endpoint"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/heartbeat/current/http-endpoint.html
applies_to:
  stack: preview
---

# Configure an HTTP endpoint for metrics [http-endpoint]


::::{warning}
This functionality is in technical preview and may be changed or removed in a future release. Elastic will work to fix any issues, but features in technical preview are not subject to the support SLA of official GA features.
::::


Heartbeat can expose internal metrics through an HTTP endpoint. These are useful to monitor the internal state of the Beat. For security reasons the endpoint is disabled by default, as you may want to avoid exposing this info.

The HTTP endpoint has the following configuration settings:

`http.enabled`
:   (Optional) Enable the HTTP endpoint. Default is `false`.

`http.host`
:   (Optional) Bind to this hostname, IP address, unix socket (unix:///var/run/heartbeat.sock) or Windows named pipe (npipe:///heartbeat). It is recommended to use only localhost. Default is `localhost`

`http.port`
:   (Optional) Port on which the HTTP endpoint will bind. Default is `5066`.

`http.named_pipe.user`
:   (Optional) User to use to create the named pipe, only work on Windows, Default to the current user.

`http.named_pipe.security_descriptor`
:   (Optional) Windows Security descriptor string defined in the SDDL format. Default to read and write permission for the current user.

`http.pprof.enabled`
:   (Optional) Enable the `/debug/pprof/` endpoints when serving HTTP. It is recommended that this is only enabled on localhost as these endpoints may leak data. Default is `false`.

`http.pprof.block_profile_rate`
:   (Optional) `block_profile_rate` controls the fraction of goroutine blocking events that are reported in the blocking profile available from `/debug/pprof/block`. The profiler aims to sample an average of one blocking event per rate nanoseconds spent blocked. To include every blocking event in the profile, pass rate = 1. To turn off profiling entirely, pass rate ⇐ 0. Defaults to 0.

`http.pprof.mem_profile_rate`
:   (Optional) `mem_profile_rate` controls the fraction of memory allocations that are recorded and reported in the memory profile available from `/debug/pprof/heap`. The profiler aims to sample an average of one allocation per `mem_profile_rate` bytes allocated. To include every allocated block in the profile, set `mem_profile_rate` to 1. To turn off profiling entirely, set `mem_profile_rate` to 0. Defaults to 524288.

`http.pprof.mutex_profile_rate`
:   (Optional) `mutex_profile_rate` controls the fraction of mutex contention events that are reported in the mutex profile available from `/debug/pprof/mutex`. On average 1/rate events are reported. To turn off profiling entirely, pass rate 0. The default value is 0.

This is the list of paths you can access. For pretty JSON output append `?pretty` to the URL.

You can query a unix socket using the `cURL` command and the `--unix-socket` flag.

```js
curl -XGET --unix-socket '/var/run/heartbeat.sock' 'http:/stats/?pretty'
```


## Info [_info]

`/` provides basic info from the Heartbeat. Example:

```js
curl -XGET 'localhost:5066/?pretty'
```

```js subs=true
{
  "beat": "heartbeat",
  "hostname": "example.lan",
  "name": "example.lan",
  "uuid": "34f6c6e1-45a8-4b12-9125-11b3e6e89866",
  "version": "{{version.stack}}"
}
```


## Stats [_stats]

`/stats` reports internal metrics. Example:

```js
curl -XGET 'localhost:5066/stats?pretty'
```

```js
{
  "beat": {
    "cpu": {
      "system": {
        "ticks": 1710,
        "time": {
          "ms": 1712
        }
      },
      "total": {
        "ticks": 3420,
        "time": {
          "ms": 3424
        },
        "value": 3420
      },
      "user": {
        "ticks": 1710,
        "time": {
          "ms": 1712
        }
      }
    },
    "info": {
      "ephemeral_id": "ab4287c4-d907-4d9d-b074-d8c3cec4a577",
      "uptime": {
        "ms": 195547
      }
    },
    "memstats": {
      "gc_next": 17855152,
      "memory_alloc": 9433384,
      "memory_total": 492478864,
      "rss": 50405376
    },
    "runtime": {
      "goroutines": 22
    }
  },
  "libbeat": {
    "config": {
      "module": {
        "running": 0,
        "starts": 0,
        "stops": 0
      },
      "scans": 1,
      "reloads": 1
    },
    "output": {
      "events": {
        "acked": 0,
        "active": 0,
        "batches": 0,
        "dropped": 0,
        "duplicates": 0,
        "failed": 0,
        "total": 0
      },
      "read": {
        "bytes": 0,
        "errors": 0
      },
      "type": "elasticsearch",
      "write": {
        "bytes": 0,
        "errors": 0
      }
    },
    "pipeline": {
      "clients": 6,
      "events": {
        "active": 716,
        "dropped": 0,
        "failed": 0,
        "filtered": 0,
        "published": 716,
        "retry": 278,
        "total": 716
      },
      "queue": {
        "acked": 0
      }
    }
  },
  "system": {
    "cpu": {
      "cores": 4
    },
    "load": {
      "1": 2.22,
      "15": 1.8,
      "5": 1.74,
      "norm": {
        "1": 0.555,
        "15": 0.45,
        "5": 0.435
      }
    }
  }
}
```

The actual output may contain more metrics specific to Heartbeat

