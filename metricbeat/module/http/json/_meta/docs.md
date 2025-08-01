This is the `json` metricset of the HTTP module.


## Features and Configuration [_features_and_configuration]

The JSON structure returned by the HTTP endpoint will be added to the provided `namespace` field as shown in the following example:

```json
{
  "@timestamp": "2017-05-01T13:00:24.745Z",
  "beat": {
    "hostname": "mbp",
    "name": "mbp",
    "version": "6.0.0-alpha1"
  },
  "http": {
    "http_json_namespace": {
      "date": "05-01-2017",
      "milliseconds_since_epoch": 1493643625474.000000,
      "time": "01:00:25 PM"
    }
  },
  "metricset": {
    "host": "date.jsontest.com",
    "module": "http",
    "name": "json",
    "namespace": "http_json_namespace",
    "rtt": 238397
  },
  "type": "metricsets"
}
```

Here the response from `date.jsontest.com` is returned in the configured `http_json_namespace` namespace:

```json
{
      "date": "05-01-2017",
      "milliseconds_since_epoch": 1493643625474.000000,
      "time": "01:00:25 PM"
}
```

It is required to set a namespace in the general module config section.


### json.is_array [_json_is_array]

With this configuration enabled the `json` metricset expects the JSON structure returned by the HTTP endpoint to be an array. Further, it creates separate events for each element in the array.


### request.enabled [_request_enabled]

With this configuration enabled additional information about the request are included. This includes the following information:

* HTTP Header
* HTTP Method
* Body/Payload

Example:

```json
{
  "@timestamp": "2017-05-01T13:00:24.745Z",
  "beat": {
    "hostname": "mbp",
    "name": "mbp",
    "version": "6.0.0-alpha1"
  },
  "http": {
    "http_json_namespace": {
      "date": "05-01-2017",
      "milliseconds_since_epoch": 1493643625474.000000,
      "time": "01:00:25 PM"
    },
    "request": {
      "body": "",
      "headers": {
        "Accept": "application/json"
      },
      "method": "GET"
    }
  },
  "metricset": {
    "host": "date.jsontest.com",
    "module": "http",
    "name": "json",
    "namespace": "http_json_namespace",
    "rtt": 238397
  },
  "type": "metricsets"
}
```


### response.enabled [_response_enabled]

With this configuration enabled additional information about the response are included. This includes the following information:

* HTTP Header
* HTTP Status Code

Example:

```json
{
  "@timestamp": "2017-05-01T13:00:24.745Z",
  "beat": {
    "hostname": "mbp",
    "name": "mbp",
    "version": "6.0.0-alpha1"
  },
  "http": {
    "http_json_namespace": {
      "date": "05-01-2017",
      "milliseconds_since_epoch": 1493643625474.000000,
      "time": "01:00:25 PM"
    },
    "response": {
      "headers": {
        "Access-Control-Allow-Origin": "*",
        "Content-Length": "100",
        "Content-Type": "application/json; charset=ISO-8859-1",
        "Date": "Mon, 01 May 2017 13:08:38 GMT",
        "Server": "Google Frontend",
        "X-Cloud-Trace-Context": "3f532d170112fc5b2a0b94fcbd6493b3"
      },
      "code": 200
    }
  },
  "metricset": {
    "host": "date.jsontest.com",
    "module": "http",
    "name": "json",
    "namespace": "http_json_namespace",
    "rtt": 238397
  },
  "type": "metricsets"
}
```


### dedot.enabled [_dedot_enabled]

With this configuration enabled dots in json field names  are replaced with `_` character,

Example:

```json
{
  "@timestamp": "2017-05-01T13:00:24.745Z",
  "beat": {
    "hostname": "mbp",
    "name": "mbp",
    "version": "6.0.0-alpha1"
  },
  "http": {
    "http_json_namespace": {
      "date": "05-01-2017",
      "milliseconds_since_epoch": 1493643625474.000000,
      "time": "01:00:25 PM"
    },
    "response": {
      "headers": {
        "Access-Control-Allow-Origin": "*",
        "Content-Length": "100",
        "Content-Type": "application/json; charset=ISO-8859-1",
        "Date": "Mon, 01 May 2017 13:08:38 GMT",
        "Server": "Google Frontend",
        "X-Cloud-Trace-Context": "3f532d170112fc5b2a0b94fcbd6493b3"
      },
      "code": 200
    }
  },
  "metricset": {
    "host": "date.jsontest.com",
    "module": "http",
    "name": "json",
    "namespace": "http_json_namespace",
    "rtt": 238397
  },
  "type": "metricsets"
}
```


## Exposed fields, Dashboards, Indexes, etc. [_exposed_fields_dashboards_indexes_etc]

Since this is a general purpose module that can be tailored for any application that exposes a JSON structure, it comes with no exposed fields description, dashboards or index patterns.
