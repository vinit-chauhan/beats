---
navigation_title: "Internal queue"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/metricbeat/current/configuring-internal-queue.html
applies_to:
  stack: ga
---

# Configure the internal queue [configuring-internal-queue]


Metricbeat uses an internal queue to store events before publishing them. The queue is responsible for buffering and combining events into batches that can be consumed by the outputs. The outputs will use bulk operations to send a batch of events in one transaction.

You can configure the type and behavior of the internal queue by setting options in the `queue` section of the `metricbeat.yml` config file or by setting options in the `queue` section of the output. Only one queue type can be configured.

This sample configuration sets the memory queue to buffer up to 4096 events:

```yaml
queue.mem:
  events: 4096
```


## Configure the memory queue [configuration-internal-queue-memory]

The memory queue keeps all events in memory.

The memory queue waits for the output to acknowledge or drop events. If the queue is full, no new events can be inserted into the memory queue. Only after the signal from the output will the queue free up space for more events to be accepted.

The memory queue is controlled by the parameters `flush.min_events` and `flush.timeout`. `flush.min_events` gives a limit on the number of events that can be included in a single batch, and `flush.timeout` specifies how long the queue should wait to completely fill an event request. If the output supports a `bulk_max_size` parameter, the maximum batch size will be the smaller of `bulk_max_size` and `flush.min_events`.

`flush.min_events` is a legacy parameter, and new configurations should prefer to control batch size with `bulk_max_size`. As of 8.13, there is never a performance advantage to limiting batch size with `flush.min_events` instead of `bulk_max_size`.

In synchronous mode, an event request is always filled as soon as events are available, even if there are not enough events to fill the requested batch. This is useful when latency must be minimized. To use synchronous mode, set `flush.timeout` to 0.

For backwards compatibility, synchronous mode can also be activated by setting `flush.min_events` to 0 or 1. In this case, batch size will be capped at 1/2 the queue capacity.

In asynchronous mode, an event request will wait up to the specified timeout to try and fill the requested batch completely. If the timeout expires, the queue returns a partial batch with all available events. To use asynchronous mode, set `flush.timeout` to a positive duration, e.g. `5s`.

This sample configuration forwards events to the output when there are enough events to fill the output’s request (usually controlled by `bulk_max_size`, and limited to at most 512 events by `flush.min_events`), or when events have been waiting for 5s without filling the requested size:

```yaml
queue.mem:
  events: 4096
  flush.min_events: 512
  flush.timeout: 5s
```


## Configuration options [_configuration_options_14]

You can specify the following options in the `queue.mem` section of the `metricbeat.yml` config file:


#### `events` [queue-mem-events-option]

Number of events the queue can store.

The default value is 3200 events.


#### `flush.min_events` [queue-mem-flush-min-events-option]

If greater than 1, specifies the maximum number of events per batch. In this case the output must wait for the queue to accumulate the requested number of events or for `flush.timeout` to expire before publishing.

If 0 or 1, sets the maximum number of events per batch to half the queue size, and sets the queue to synchronous mode (equivalent to `flush.timeout` of 0).

The default value is 1600.


#### `flush.timeout` [queue-mem-flush-timeout-option]

Maximum wait time for event requests from the output to be fulfilled. If set to 0s, events are returned immediately.

The default value is 10s.


## Configure the disk queue [configuration-internal-queue-disk]

The disk queue stores pending events on the disk rather than main memory. This allows Beats to queue a larger number of events than is possible with the memory queue, and to save events when a Beat or device is restarted. This increased reliability comes with a performance tradeoff, as every incoming event must be written and read from the device’s disk. However, for setups where the disk is not the main bottleneck, the disk queue gives a simple and relatively low-overhead way to add a layer of robustness to incoming event data.

To enable the disk queue with default settings, specify a maximum size:

```yaml
queue.disk:
  max_size: 10GB
```

The queue will use up to the specified maximum size on disk. It will only use as much space as required. For example, if the queue is only storing 1GB of events, then it will only occupy 1GB on disk no matter how high the maximum is. Queue data is deleted from disk after it has been successfully sent to the output.


### Configuration options [configuration-internal-queue-disk-reference]

You can specify the following options in the `queue.disk` section of the `metricbeat.yml` config file:


#### `path` [_path]

The path to the directory where the disk queue should store its data files. The directory is created on startup if it doesn’t exist.

The default value is `"${path.data}/diskqueue"`.


#### `max_size` (required) [_max_size_required]

The maximum size the queue should use on disk. Events that exceed this maximum will either pause their input or be discarded, depending on the input’s configuration.

A value of `0` means that no maximum size is enforced, and the queue can grow up to the amount of free space on the disk. This value should be used with caution, as completely filling a system’s main disk can make it inoperable. It is best to use this setting only with a dedicated data or backup partition that will not interfere with Metricbeat or the rest of the host system.

The default value is `10GB`.


#### `segment_size` [_segment_size]

Data added to the queue is stored in segment files. Each segment contains some number of events waiting to be sent to the outputs, and is deleted when all its events are sent. By default, segment size is limited to 1/10 of the maximum queue size. Using a smaller size means that the queue will use more data files, but they will be deleted more quickly after use. Using a larger size means some data will take longer to delete, but the queue will use fewer auxiliary files. It is usually fine to leave this value unchanged.

The default value is `max_size / 10`.


#### `read_ahead` [_read_ahead]

The number of events that should be read from disk into memory while waiting for an output to request them. If you find outputs are slowing down because they can’t read as many events at a time, adjusting this setting upward may help, at the cost of higher memory usage.

The default value is `512`.


#### `write_ahead` [_write_ahead]

The number of events the queue should accept and store in memory while waiting for them to be written to disk. If you find the queue’s memory use is too high because events are waiting too long to be written to disk, adjusting this setting downward may help, at the cost of reduced event throughput. On the other hand, if inputs are waiting or discarding events because they are being produced faster than the disk can handle, adjusting this setting upward may help, at the cost of higher memory usage.

The default value is `2048`.


#### `retry_interval` [_retry_interval]

Some disk errors may block operation of the queue, for example a permission error writing to the data directory, or a disk full error while writing an event. In this case, the queue reports the error and retries after pausing for the time specified in `retry_interval`.

The default value is `1s` (one second).


#### `max_retry_interval` [_max_retry_interval]

When there are multiple consecutive errors writing to the disk, the queue increases the retry interval by factors of 2 up to a maximum of `max_retry_interval`. Increase this value if you are concerned about logging too many errors or overloading the host system if the target disk becomes unavailable for an extended time.

The default value is `30s` (thirty seconds).

