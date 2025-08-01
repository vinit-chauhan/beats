---
navigation_title: "CEL"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/filebeat/current/filebeat-input-cel.html
sub:
  mito_docs: https://pkg.go.dev/github.com/elastic/mito
  mito_version: v1.22.0
applies_to:
  stack: ga 8.6.0
---

# Common Expression Language input [filebeat-input-cel]

Use the `cel` input to read messages from a file path or HTTP API with a variety of payloads using the [Common Expression Language (CEL)](https://opensource.google.com/projects/cel) and the [mito]({{mito_docs}}/lib) CEL extension libraries.

CEL is a non-Turing complete language that can perform evaluation of expression in inputs, which can include file and API endpoints using the mito extension library. The `cel` input periodically runs a CEL program that is given an execution environment that may be configured by the user, and publishes the set of events that result from the program evaluation. Optionally the CEL program may return cursor states that will be provided to the next execution of the CEL program. The cursor states may be used to control the behavior of the program.

This input supports:

* Auth

    * Basic
    * Digest
    * OAuth2

* Retrieval at a configurable interval
* Pagination
* Retries
* Rate limiting
* Proxying

Example configurations:

```yaml
filebeat.inputs:
# Fetch your public IP every minute.
- type: cel
  interval: 1m
  resource.url: https://api.ipify.org/?format=json
  program: |
    get(state.url).Body.as(body, {
        "events": [body.decode_json()]
    })
```

or equivalently using the text format from ipify.org

```yaml
filebeat.inputs:
# Fetch your public IP every minute.
- type: cel
  interval: 1m
  resource.url: https://api.ipify.org/?format=text
  program: |
    {
        "events": [{"ip": string(get(state.url).Body)}]
    }
```

```yaml
filebeat.inputs:
- type: cel
  resource.url: http://localhost:9200/_search
  state:
    scroll: 5m
  program: |
    (
        !has(state.cursor) || !has(state.cursor.scroll_id) ?
            post(state.url+"?scroll=5m", "", "")
        :
            post(
                state.url+"/scroll?"+{"scroll_id": [state.cursor.scroll_id]}.format_query(),
                "application/json",
                {"scroll": state.scroll}.encode_json()
            )
    ).as(resp, resp.Body.decode_json().as(body, {
            "events": body.hits.hits,
            "cursor": {"scroll_id": body._scroll_id},
    }))
```

## Execution [_execution]

The execution environment provided for the input includes the functions, macros, and global variables provided by the mito library. A single JSON object is provided as an input accessible through a `state` variable. `state` contains a string `url` field and may contain arbitrary other fields configured via the input’s `state` configuration. If the CEL program saves cursor states between executions of the program, the configured `state.cursor` value will be replaced by the saved cursor prior to execution.

On start the `state` is will be something like this:

```json
{
    "url": <resource address>,
    "cursor": { ... },
    ...
}
```

The `state.url` field will be present and may be an HTTP end-point or a file path. It is the responsibility of the CEL program to handle removing the scheme from a file URL if it is present. The `state.url` field may be mutated during execution of the program, but the mutated state will not be persisted between restarts The `state.url` field must be present in the returned value to ensure that it is available in the next evaluation unless the program has the resource address hard-coded in or it is available from the cursor.

Additional fields may be present at the root of the object and if the program tolerates it, the cursor value may be absent. Only the cursor is persisted over restarts, but all fields in state are retained between iterations of the processing loop except for the produced events array, see below.

If the cursor is present the program should perform and process requests based on its value. If cursor is not present the program must have alternative logic to determine what requests to make.

After completion of a program’s execution it should return a single object with a structure looking like this:

```json
{
    "events": [ <1>
        {...},
        ...
    ],
    "cursor": [ <2>
        {...},
        ...
    ],
    "url": <resource address>,
    "status_code": <HTTP request status code if a network request>,
    "header": <HTTP response headers if a network request>,
    "rate_limit": <HTTP rate limit map if required by API>, <3>
    "want_more": false <4>
}
```

1. The `events` field must be present, but may be empty or null. If it is not empty, it must only have objects as elements. The field should be an array, but in the case of an error condition in the CEL program it is acceptable to return a single object instead of an array; this will will be wrapped as an array for publication and an error will be logged. If the single object contains a key, "error", the error value will be used to update the status of the input to report to Elastic Agent. This can be used to more rapidly respond to API failures. It is recommended that the object conforms to ECS field definitions, but this is not enforced.
2. If `cursor` is present it must be either be a single object or an array with the same length as events; each element *i* of the `cursor` will be the details for obtaining the events at and beyond event *i* in the `events` array. If the `cursor` is a single object it is will be the details for obtaining events after the last event in the `events` array and will only be retained on successful publication of all the events in the `events` array.
3. If `rate_limit` is present it must be a map with numeric fields `rate` and `burst`. The `rate_limit` field may also have a string `error` field and other fields which will be logged. If it has an `error` field, the `rate` and `burst` will not be used to set rate limit behavior. The [Limit]({{mito_docs}}@{{mito_version}}/lib#Limit), and [Okta Rate Limit policy]({{mito_docs}}@{{mito_version}}/lib#OktaRateLimit) and [Draft Rate Limit policy]({{mito_docs}}@{{mito_version}}/lib#DraftRateLimit) documentation show how to construct this field.
4. The evaluation is repeated with the new state, after removing the events field, if the "want_more" field is present and true, and a non-zero events array is returned. If the "want_more" field is present after a failed evaluation, it is set to false.


The `status_code`, `header` and `rate_limit` values may be omitted if the program is not interacting with an HTTP API end-point and so will not be needed to contribute to program control.


## Debug state logging [_debug_state_logging]

The CEL input will log the complete state after evaluation when logging at the DEBUG level. This will include any sensitive or secret information kept in the `state` object, and so DEBUG level logging should not be used in production when sensitive information is retained in the `state` object. See [`redact`](#cel-state-redact) configuration parameters for settings to exclude sensitive fields from DEBUG logs.


## CEL extension libraries [_cel_extension_libraries]

As noted above the `cel` input provides functions, macros, and global variables to extend the language.

* [AWS v4 request signing]({{mito_docs}}@{{mito_version}}/lib#AWS) {applies_to}`stack: ga 8.19.0, unavailable 9.0.0, ga 9.1.0`

    * [Sign AWS from env]({{mito_docs}}@{{mito_version}}/lib#hdr-Sign_AWS_from_env-AWS)
    * [Sign AWS from shared credentials]({{mito_docs}}@{{mito_version}}/lib#hdr-Sign_AWS_from_shared_credentials-AWS)
    * [Sign AWS from static credentials]({{mito_docs}}@{{mito_version}}/lib#hdr-Sign_AWS_from_static_credentials-AWS)

* [Collections]({{mito_docs}}@{{mito_version}}/lib#Collections) {applies_to}`stack: ga 8.6.0`

    * [Collate]({{mito_docs}}@{{mito_version}}/lib#hdr-Collate-Collections)
    * [Drop]({{mito_docs}}@{{mito_version}}/lib#hdr-Drop-Collections)
    * [Drop Empty]({{mito_docs}}@{{mito_version}}/lib#hdr-Drop_Empty-Collections)
    * [Flatten]({{mito_docs}}@{{mito_version}}/lib#hdr-Flatten-Collections)
    * [Front]({{mito_docs}}@{{mito_version}}/lib#hdr-Front-Collections) {applies_to}`stack: ga 8.18.0`
    * [Keys]({{mito_docs}}@{{mito_version}}/lib#hdr-Keys-Collections) {applies_to}`stack: ga 8.13.0`
    * [Max]({{mito_docs}}@{{mito_version}}/lib#hdr-Max-Collections)
        * list maximum
        * pair maximum {applies_to}`stack: ga 8.18.0`
    * [Min]({{mito_docs}}@{{mito_version}}/lib#hdr-Min-Collections)
        * list minimum
        * pair minimum {applies_to}`stack: ga 8.18.0`
    * [Sum]({{mito_docs}}@{{mito_version}}/lib#hdr-Sum-Collections) {applies_to}`stack: ga 8.18.0`
    * [Tail]({{mito_docs}}@{{mito_version}}/lib#hdr-Tail-Collections)
        * one parameter {applies_to}`stack: ga 8.15.0`
        * two parameter {applies_to}`stack: ga 8.18.0`
    * [Values]({{mito_docs}}@{{mito_version}}/lib#hdr-Values-Collections) {applies_to}`stack: ga 8.13.0`
    * [With]({{mito_docs}}@{{mito_version}}/lib#hdr-With-Collections)
    * [With Replace]({{mito_docs}}@{{mito_version}}/lib#hdr-With_Replace-Collections)
    * [With Update]({{mito_docs}}@{{mito_version}}/lib#hdr-With_Update-Collections)
    * [Zip]({{mito_docs}}@{{mito_version}}/lib#hdr-Zip-Collections) {applies_to}`stack: ga 8.9.2`

* [Crypto]({{mito_docs}}@{{mito_version}}/lib#Crypto) {applies_to}`stack: ga 8.6.0`

    * [Base64]({{mito_docs}}@{{mito_version}}/lib#hdr-Base64-Crypto)
    * [Base64 Decode]({{mito_docs}}@{{mito_version}}/lib#hdr-Base64_Decode-Crypto) {applies_to}`stack: ga 8.14.0`
    * [Base64 Raw]({{mito_docs}}@{{mito_version}}/lib#hdr-Base64_Raw-Crypto)
    * [Base64 Raw Decode]({{mito_docs}}@{{mito_version}}/lib#hdr-Base64_Raw_Decode-Crypto) {applies_to}`stack: ga 8.14.0`
    * [Hex]({{mito_docs}}@{{mito_version}}/lib#hdr-Hex-Crypto)
    * [Hex Decode]({{mito_docs}}@{{mito_version}}/lib#hdr-Hex_Decode-Crypto) {applies_to}`stack: ga 8.18.1`
    * [MD5]({{mito_docs}}@{{mito_version}}/lib#hdr-MD5-Crypto)
    * [SHA-1]({{mito_docs}}@{{mito_version}}/lib#hdr-SHA_1-Crypto)
    * [SHA-256]({{mito_docs}}@{{mito_version}}/lib#hdr-SHA_256-Crypto)
    * [HMAC]({{mito_docs}}@{{mito_version}}/lib#hdr-HMAC-Crypto)
    * [UUID]({{mito_docs}}@{{mito_version}}/lib#hdr-UUID-Crypto)

* [File]({{mito_docs}}@{{mito_version}}/lib#File) — the file extension is initialized with MIME handlers for "application/gzip", ["application/x-ndjson"]({{mito_docs}}@{{mito_version}}/lib#NDJSON), ["application/zip"]({{mito_docs}}@{{mito_version}}/lib#Zip), ["text/csv; header=absent"]({{mito_docs}}@{{mito_version}}/lib#CSVNoHeader), and ["text/csv; header=present"]({{mito_docs}}@{{mito_version}}/lib#CSVHeader). {applies_to}`stack: ga 8.6.0`

    * [Dir]({{mito_docs}}@{{mito_version}}/lib#hdr-Dir-File)
    * [File]({{mito_docs}}@{{mito_version}}/lib#hdr-File-File)

* [HTTP]({{mito_docs}}@{{mito_version}}/lib#HTTP) {applies_to}`stack: ga 8.6.0`

    * [HEAD]({{mito_docs}}@{{mito_version}}/lib#hdr-HEAD-HTTP)
    * [GET]({{mito_docs}}@{{mito_version}}/lib#hdr-GET-HTTP)
    * [GET Request]({{mito_docs}}@{{mito_version}}/lib#hdr-GET_Request-HTTP)
    * [POST]({{mito_docs}}@{{mito_version}}/lib#hdr-POST-HTTP)
    * [POST Request]({{mito_docs}}@{{mito_version}}/lib#hdr-POST_Request-HTTP)
    * [Request]({{mito_docs}}@{{mito_version}}/lib#hdr-Request-HTTP)
    * [Basic Authentication]({{mito_docs}}@{{mito_version}}/lib#hdr-Basic_Authentication-HTTP) {applies_to}`stack: ga 8.7.0`
    * [Do Request]({{mito_docs}}@{{mito_version}}/lib#hdr-Do_Request-HTTP)
    * [Parse URL]({{mito_docs}}@{{mito_version}}/lib#hdr-Parse_URL-HTTP)
    * [Format URL]({{mito_docs}}@{{mito_version}}/lib#hdr-Format_URL-HTTP)
    * [Parse Query]({{mito_docs}}@{{mito_version}}/lib#hdr-Parse_Query-HTTP)
    * [Format Query]({{mito_docs}}@{{mito_version}}/lib#hdr-Format_Query-HTTP)

* [JSON]({{mito_docs}}@{{mito_version}}/lib#JSON) {applies_to}`stack: ga 8.6.0`

    * [Encode JSON]({{mito_docs}}@{{mito_version}}/lib#hdr-Encode_JSON-JSON)
    * [Decode JSON]({{mito_docs}}@{{mito_version}}/lib#hdr-Decode_JSON-JSON)
    * [Decode JSON Stream]({{mito_docs}}@{{mito_version}}/lib#hdr-Decode_JSON_Stream-JSON)

* [XML]({{mito_docs}}@{{mito_version}}/lib#XML) — the XML extension is initialized with XML schema definitions provided via the `xsd` configuration option.  {applies_to}`stack: ga 8.9.0`

    * [Decode XML]({{mito_docs}}@{{mito_version}}/lib#hdr-Decode_XML-XML)
        * Optional XSD definition in one-parameter form. {applies_to}`stack: ga 8.18.1`

* [Limit]({{mito_docs}}@{{mito_version}}/lib#Limit) — the rate limit extension is initialized with [Okta (as "okta")]({{mito_docs}}@{{mito_version}}/lib#OktaRateLimit) and the [Draft Rate Limit (as "draft")]({{mito_docs}}@{{mito_version}}/lib#DraftRateLimit) policies. {applies_to}`stack: ga 8.6.0`

    * [Rate Limit]({{mito_docs}}@{{mito_version}}/lib#hdr-Rate_Limit-Limit)

* [MIME]({{mito_docs}}@{{mito_version}}/lib#MIME) — the MIME extension is initialized with MIME handlers for "application/gzip", ["application/x-ndjson"]({{mito_docs}}@{{mito_version}}/lib#NDJSON), ["application/zip"]({{mito_docs}}@{{mito_version}}/lib#Zip), ["text/csv; header=absent"]({{mito_docs}}@{{mito_version}}/lib#CSVNoHeader), and ["text/csv; header=present"]({{mito_docs}}@{{mito_version}}/lib#CSVHeader). {applies_to}`stack: ga 8.6.0`

    * [MIME]({{mito_docs}}@{{mito_version}}/lib#hdr-MIME-MIME)

* [Regexp]({{mito_docs}}@{{mito_version}}/lib#Regexp) — the regular expression extension is initialized with the patterns specified in the user input configuration via the `regexp` field. {applies_to}`stack: ga 8.6.0`

    * [RE Match]({{mito_docs}}@{{mito_version}}/lib#hdr-RE_Match)
    * [RE Find]({{mito_docs}}@{{mito_version}}/lib#hdr-RE_Find)
    * [RE Find All]({{mito_docs}}@{{mito_version}}/lib#hdr-RE_Find_All)
    * [RE Find Submatch]({{mito_docs}}@{{mito_version}}/lib#hdr-RE_Find_Submatch)
    * [RE Find All Submatch]({{mito_docs}}@{{mito_version}}/lib#hdr-RE_Find_All_Submatch)
    * [RE Replace All]({{mito_docs}}@{{mito_version}}/lib#hdr-RE_Replace_All)

* [Printf]({{mito_docs}}@{{mito_version}}/lib#Printf)

    * [Sprintf]({{mito_docs}}@{{mito_version}}/lib#hdr-Sprintf-Printf)

* [Strings]({{mito_docs}}@{{mito_version}}/lib#Strings) {applies_to}`stack: ga 8.7.0`

    * [String Methods]({{mito_docs}}@{{mito_version}}/lib#hdr-String_Methods-Strings)
    * [String List Methods]({{mito_docs}}@{{mito_version}}/lib#hdr-String_List_Methods-Strings)
    * [Bytes Methods]({{mito_docs}}@{{mito_version}}/lib#hdr-Bytes_Methods-Strings) {applies_to}`stack: ga 8.15.0`

* [Time]({{mito_docs}}@{{mito_version}}/lib#Time) {applies_to}`stack: ga 8.6.0`

    * [Format]({{mito_docs}}@{{mito_version}}/lib#hdr-Format-Time)
    * [Parse Time]({{mito_docs}}@{{mito_version}}/lib#hdr-Parse_Time-Time)
    * [Round]({{mito_docs}}@{{mito_version}}/lib#hdr-Round-Time) {applies_to}`stack: ga 8.19.0, unavailable 9.0.0, ga 9.1.0`
    * [Global Variables]({{mito_docs}}@{{mito_version}}/lib#hdr-Global_Variables-Time)
        * Support for [`DateOnly`](https://pkg.go.dev/time#DateOnly), [`DateTime`](https://pkg.go.dev/time#DateTime) and [`TimeOnly`](https://pkg.go.dev/time#TimeOnly) time formats. {applies_to}`stack: ga 8.15.0`

* [Try]({{mito_docs}}@{{mito_version}}/lib#Try) {applies_to}`stack: ga 8.6.0`

    * [Try]({{mito_docs}}@{{mito_version}}/lib#hdr-Try-Try)
    * [Is Error]({{mito_docs}}@{{mito_version}}/lib#hdr-Is_Error-Try)

* [Debug]({{mito_docs}}@{{mito_version}}/lib#Debug) — the debug handler registers a logger with the name extension `cel_debug` and calls to the CEL `debug` function are emitted to that logger. {applies_to}`stack: ga 8.10.3`

    * [Debug]({{mito_docs}}@{{mito_version}}/lib#hdr-Debug)


In addition to the extensions provided in the packages listed above, a global variable `useragent` is also provided which gives the user CEL program access to the filebeat user-agent string. By default, this value is assigned to all requests' user-agent headers unless the CEL program has already set the user-agent header value. Programs wishing to not provide a user-agent, should set this header to the empty string, `""`.

Host environment variables are made available via the global map `env`. Only environment variables that have been allow listed via the `allowed_environment` configuration list are visible to the CEL program.

The CEL environment enables the [optional types](https://pkg.go.dev/github.com/google/cel-go/cel#OptionalTypes) library using the version defined [here]({{mito_docs}}@{{mito_version}}/lib#OptionalTypesVersion) and the [two-variable comprehensions extensions](https://pkg.go.dev/github.com/google/cel-go/ext#TwoVarComprehensions) library using the version defined [here]({{mito_docs}}@{{mito_version}}/lib#TwoVarComprehensionVersion).

* Optional types {applies_to}`stack: ga 8.12.0`
* Two-variable comprehensions {applies_to}`stack: ga 8.19.0, unavailable 9.0.0, ga 9.1.0`

Additionally, it supports authentication via:

* Basic Authentication
* Digest Authentication {applies_to}`stack: ga 8.12.0`
* OAuth2
* token authentication {applies_to}`stack: ga 8.19.0, unavailable 9.0.0, ga 9.1.0`

Example configurations with authentication:

```yaml
filebeat.inputs:
- type: cel
  auth.basic:
    user: user@domain.tld
    password: P@$$W0₹D
  resource.url: http://localhost
```

```yaml
filebeat.inputs:
- type: cel
  auth.digest:
    user: user@domain.tld
    password: P@$$W0₹D
  resource.url: http://localhost
```

```yaml
filebeat.inputs:
- type: cel
  auth.oauth2:
    client.id: 12345678901234567890abcdef
    client.secret: abcdef12345678901234567890
    token_url: http://localhost/oauth2/token
  resource.url: http://localhost
```

```yaml
filebeat.inputs:
- type: cel
  auth.oauth2:
    client.id: 12345678901234567890abcdef
    client.secret: abcdef12345678901234567890
    token_url: http://localhost/oauth2/token
    user: user@domain.tld
    password: P@$$W0₹D
  resource.url: http://localhost
```

```yaml
filebeat.inputs:
- type: cel
  auth.token:
    type: Bearer
    value: supersecret_bearer_token
  resource.url: http://localhost
```

```yaml
filebeat.inputs:
- type: cel
  auth.token:
    type: Token
    value: supersecret_token
  resource.url: http://localhost
```

## Input state [input-state-cel]

The `cel` input keeps a runtime state between requests. This state can be accessed by the CEL program and may contain arbitrary objects.

The state must contain a `url` string and may contain any object the user wishes to store in it.

All objects are stored at runtime, except `cursor`, which has values that are persisted between restarts.


## CEL input and handling numbers [_cel_input_and_numbers]

Numeric values passed in to and out of a CEL evaluation environment are passed as floating point
values. This can sometimes cause issues when the numbers in the input state are expected to be
integers, and may result in unexpected field values being ingested into Elasticsearch documents when
other parts of the ingest pipeline render floating point values with E-notation or add decimal
points to numbers that are expected to be integers. This is most likely to happen when numbers are
large (not within ±10^7^). Above the maximum exact integer representation threshold for double
precision floating point values, within ±2^53^ (±9×10^15^), integer values will lose precision when
they are returned from the CEL evaluation environment. The CEL input will automatically convert
integers outside of the ±2^53^ range to strings in order to prevent loss of precision in these
values, but potentially leading to a situation where some numbers received by the ingest pipeline
are numbers and some are strings. To avoid these issues, when you have large integer values as part
of an evaluation result, convert the field value to a string before returning it, and convert input
numbers to integers explicitly at the start of a CEL program.


## Configuration options [_configuration_options_3]

The `cel` input supports the following configuration options plus the [Common options](#filebeat-input-cel-common-options) described later.


### `interval` [_interval]

Duration between repeated requests. It may make additional pagination requests in response to the initial request if pagination is enabled. Default: `60s`.


### `program` [program-cel]

The CEL program that is executed each polling period. This field is required.


### `max_executions` [max_executions-cel]

`max_executions` is the maximum number of times a CEL program can request to be re-run with a `want_more` field. This is used to ensure that accidental infinite loops do not halt processing. When the execution budget is exceeded, execution will be restarted at the next interval and a warning will be written into the logs. Default: 1000.


### `state` [state-cel]

`state` is an optional object that is passed to the CEL program as the `state` variable on the first execution. Subsequent executions of the program during the life of the input will populate the `state` variable with the return value of the previous execution, but with the `state.events` field removed. Except for the `state.cursor` field, returned `state` data does not persist over restarts.


### `state.cursor` [cursor-cel]

The cursor is an object available as `state.cursor` where arbitrary values may be stored. Cursor state is kept between input restarts and updated after each event of a request has been published. When a cursor is used the CEL program must either create a cursor state for each event that is returned by the program, or a single cursor that reflect the cursor for completion of the full set of events.

```yaml
filebeat.inputs:
# Fetch your public IP every minute and note when the last request was made.
- type: cel
  interval: 1m
  resource.url: https://api.ipify.org/?format=json
  program: |
    get(state.url).Body.as(body, {
        "events": [body.decode_json().with({
            "last_requested_at": has(state.cursor) && has(state.cursor.last_requested_at) ?
                state.cursor.last_requested_at
            :
                now
        })],
        "cursor": {"last_requested_at": now}
    })
```


### `allowed_environment` [environ-cel]

```{applies_to}
stack: ga 8.16.0
```

A list of host environment variable that will be made visible to the CEL execution environment. By default, no environment variables are visible.

```yaml
filebeat.inputs:
# Publish the list of files in $PATH every minute.
- type: cel
  interval: 1m
  resource.url: ""
  allowed_environment:
    - PATH
  program: |
{
  "events": {
    "message": env.?PATH.orValue("").split(":")
      .map(p, try(dir(p)))
      .filter(d, type(d) != type(""))
      .flatten()
      .collate("name")
  }
}
```


### `regexp` [regexp-cel]

A set of named regular expressions that may be used during a CEL program’s execution using the `regexp` extension library. The syntax used for the regular expressions is [RE2](https://github.com/google/re2/wiki/Syntax).

```yaml
filebeat.inputs:
- type: cel
  # Define two regular expressions, 'products' and 'solutions' for use during CEL execution.
  regexp:
    products: '(?i)(Elasticsearch|Beats|Logstash|Kibana)'
    solutions: '(?i)(Search|Observability|Security)'
```


### `xsd` [xsd-cel]

```{applies_to}
stack: ga 8.9.0
```

XML documents may require additional type information to enable correct parsing and ingestion. This information can be provided as an XML Schema Definitions (XSD) for XML documents using the `xsd` option. The key under which the XSD information is provided is accessed via the `decode_xml` CEL extension.

```yaml
filebeat.inputs:
- type: cel
  # Provide an XSD, 'order', for use during CEL execution (truncated for example).
  xsd:
    order: |
       <xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
         <xs:element name="order">
           <xs:complexType>
             <xs:sequence>
               <xs:element name="sender" type="xs:string"/>
               <xs:element name="address">
                 <xs:complexType>
                   <xs:sequence>
                     <xs:element name="name" type="xs:string"/>
                     <xs:element name="company" type="xs:string"/>
```

The `xsd` for an XML document structure may be omitted. {applies_to}`stack: ga 8.18.1`


### `auth.basic.enabled` [_auth_basic_enabled]

When set to `false`, disables the basic auth configuration. Default: `true`.

::::{note}
Basic auth settings are disabled if either `enabled` is set to `false` or the `auth.basic` section is missing.
::::



### `auth.basic.user` [_auth_basic_user]

The user to authenticate with.


### `auth.basic.password` [_auth_basic_password]

The password to use.


### `auth.digest.enabled` [_auth_digest_enabled]

```{applies_to}
stack: ga 8.12.0
```

When set to `false`, disables the digest auth configuration. Default: `true`.

::::{note}
digest auth settings are disabled if either `enabled` is set to `false` or the `auth.digest` section is missing.
::::



### `auth.digest.user` [_auth_digest_user]

```{applies_to}
stack: ga 8.12.0
```

The user to authenticate with.


### `auth.digest.password` [_auth_digest_password]

```{applies_to}
stack: ga 8.12.0
```

The password to use.


### `auth.digest.no_reuse` [_auth_digest_no_reuse]

```{applies_to}
stack: ga 8.12.0
```

When set to `true`, Digest Authentication challenges are not reused.


### `auth.oauth2.enabled` [_auth_oauth2_enabled]

When set to `false`, disables the oauth2 configuration. Default: `true`.

::::{note}
OAuth2 settings are disabled if either `enabled` is set to `false` or the `auth.oauth2` section is missing.
::::



### `auth.oauth2.provider` [_auth_oauth2_provider]

Used to configure supported oauth2 providers. Each supported provider will require specific settings. It is not set by default. Supported providers are: `azure`, `google`, `okta`.


### `auth.oauth2.client.id` [_auth_oauth2_client_id]

The client ID used as part of the authentication flow. It is always required except if using `google` as provider. Required for providers: `default`, `azure`, `okta`.


### `auth.oauth2.client.secret` [_auth_oauth2_client_secret]

The client secret used as part of the authentication flow. It is always required except if using `google` or `okta` as provider. Required for providers: `default`, `azure`.


### `auth.oauth2.user` [_auth_oauth2_user]

The user used as part of the authentication flow. It is required for authentication - grant type password. It is only available for provider `default`.


### `auth.oauth2.password` [_auth_oauth2_password]

The password used as part of the authentication flow. It is required for authentication - grant type password. It is only available for provider `default`.

::::{note}
user and password are required for grant_type password. If user and password is not used then it will automatically use the `token_url` and `client credential` method.
::::



### `auth.oauth2.scopes` [_auth_oauth2_scopes]

A list of scopes that will be requested during the oauth2 flow. It is optional for all providers.


### `auth.oauth2.token_url` [_auth_oauth2_token_url]

The endpoint that will be used to generate the tokens during the oauth2 flow. It is required if no provider is specified.

::::{note}
For `azure` provider either `token_url` or `azure.tenant_id` is required.
::::



### `auth.oauth2.endpoint_params` [_auth_oauth2_endpoint_params]

Set of values that will be sent on each request to the `token_url`. Each param key can have multiple values. Can be set for all providers except `google`.

```yaml
- type: cel
  auth.oauth2:
    endpoint_params:
      Param1:
        - ValueA
        - ValueB
      Param2:
        - Value
```


### `auth.oauth2.azure.tenant_id` [_auth_oauth2_azure_tenant_id]

Used for authentication when using `azure` provider. Since it is used in the process to generate the `token_url`, it can’t be used in combination with it. It is not required.

For information about where to find it, you can refer to [https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal](https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal).


### `auth.oauth2.azure.resource` [_auth_oauth2_azure_resource]

The accessed WebAPI resource when using `azure` provider. It is not required.


### `auth.oauth2.google.credentials_file` [_auth_oauth2_google_credentials_file]

The credentials file for Google.

::::{note}
Only one of the credentials settings can be set at once. If none is provided, loading default credentials from the environment will be attempted via ADC. For more information about how to provide Google credentials, please refer to [https://cloud.google.com/docs/authentication](https://cloud.google.com/docs/authentication).
::::



### `auth.oauth2.google.credentials_json` [_auth_oauth2_google_credentials_json]

Your credentials information as raw JSON.

::::{note}
Only one of the credentials settings can be set at once. If none is provided, loading default credentials from the environment will be attempted via ADC. For more information about how to provide Google credentials, please refer to [https://cloud.google.com/docs/authentication](https://cloud.google.com/docs/authentication).
::::



### `auth.oauth2.google.jwt_file` [_auth_oauth2_google_jwt_file]

The JWT Account Key file for Google.

::::{note}
Only one of the credentials settings can be set at once. If none is provided, loading default credentials from the environment will be attempted via ADC. For more information about how to provide Google credentials, please refer to [https://cloud.google.com/docs/authentication](https://cloud.google.com/docs/authentication).
::::



### `auth.oauth2.google.jwt_json` [_auth_oauth2_google_jwt_json]

The JWT Account Key file as raw JSON.

::::{note}
Only one of the credentials settings can be set at once. If none is provided, loading default credentials from the environment will be attempted via ADC. For more information about how to provide Google credentials, please refer to [https://cloud.google.com/docs/authentication](https://cloud.google.com/docs/authentication).
::::



### `auth.oauth2.google.delegated_account` [_auth_oauth2_google_delegated_account]

Email of the delegated account used to create the credentials (usually an admin). Used in combination with `auth.oauth2.google.jwt_file` or `auth.oauth2.google.jwt_json`.


### `auth.oauth2.okta.jwk_file` [_auth_oauth2_okta_jwk_file]

The RSA JWK Private Key file for your Okta Service App which is used for interacting with Okta Org Auth Server to mint tokens with okta.* scopes.

::::{note}
Only one of the credentials settings can be set at once. For more information please refer to [https://developer.okta.com/docs/guides/implement-oauth-for-okta-serviceapp/main/](https://developer.okta.com/docs/guides/implement-oauth-for-okta-serviceapp/main/)
::::



### `auth.oauth2.okta.jwk_json` [_auth_oauth2_okta_jwk_json]

The RSA JWK Private Key JSON for your Okta Service App which is used for interacting with Okta Org Auth Server to mint tokens with okta.* scopes.

::::{note}
Only one of the credentials settings can be set at once. For more information please refer to [https://developer.okta.com/docs/guides/implement-oauth-for-okta-serviceapp/main/](https://developer.okta.com/docs/guides/implement-oauth-for-okta-serviceapp/main/)
::::



### `auth.oauth2.okta.jwk_pem` [_auth_oauth2_okta_jwk_pem]

The RSA JWK private key PEM block for your Okta Service App which is used for interacting with Okta Org Auth Server to mint tokens with okta.* scopes.

::::{note}
Only one of the credentials settings can be set at once. For more information please refer to [https://developer.okta.com/docs/guides/implement-oauth-for-okta-serviceapp/main/](https://developer.okta.com/docs/guides/implement-oauth-for-okta-serviceapp/main/)
::::



### `auth.token.enabled` [_auth_token_enabled]

When set to `false`, disables the token authentication configuration. Default: `true`.

::::{note}
Token authentication settings are disabled if either `enabled` is set to `false` or the `auth.token` section is missing.
::::



### `auth.token.type` [_auth_token_type]

The type of token to authenticate with, for example "Token" or "Bearer".


### `auth.token.value` [_auth_token_value]

The token value to use.



### `resource.url` [resource-parameters]

The URL of the HTTP API. Required.

The API endpoint may be accessed via unix socket and Windows named pipes by adding  `+unix` or `+npipe` to the URL scheme, for example, `http+unix:///var/socket/`.


### `resource.headers` [_resource_headers]

```{applies_to}
stack: ga 8.18.1
```

Headers to be added to all requests. Headers are added before authentication headers, so any collision between headers in this configuration and authentication headers will result in the colliding headers here not being included in requests. Header values must be provided as an array.

```yaml
filebeat.inputs:
# Fetch your public IP every minute.
- type: cel
  interval: 1m
  resource.url: https://api.ipify.org/?format=text
  resource.headers:
    Custom-Header:
      - Value
    Other-Custom-Header:
      - Other value
  resource.proxy_url: http://proxy.example:8080
  program: |
    {
        "events": [{"ip": string(get(state.url).Body)}]
    }
```

### `resource.timeout` [_resource_timeout]

Duration before declaring that the HTTP client connection has timed out. Valid time units are `ns`, `us`, `ms`, `s`, `m`, `h`. Default: `30s`.


### `resource.ssl` [_resource_ssl]

This specifies SSL/TLS configuration. If the ssl section is missing, the host’s CAs are used for HTTPS connections. See [SSL](/reference/filebeat/configuration-ssl.md) for more information.


### `resource.proxy_url` [_resource_proxy_url]

This specifies proxy configuration in the form of `http[s]://<user>:<password>@<server name/ip>:<port>`. Proxy headers may be configured using the `resource.proxy_headers` field which accepts a set of key/value pairs.

```yaml
filebeat.inputs:
# Fetch your public IP every minute.
- type: cel
  interval: 1m
  resource.url: https://api.ipify.org/?format=text
  resource.proxy_url: http://proxy.example:8080
  program: |
    {
        "events": [{"ip": string(get(state.url).Body)}]
    }
```


### `resource.keep_alive.disable` [_resource_keep_alive_disable]

This specifies whether to disable keep-alives for HTTP end-points. Default: `true`.


### `resource.keep_alive.max_idle_connections` [_resource_keep_alive_max_idle_connections]

The maximum number of idle connections across all hosts. Zero means no limit. Default: `0`.


### `resource.keep_alive.max_idle_connections_per_host` [_resource_keep_alive_max_idle_connections_per_host]

The maximum idle connections to keep per-host. If zero, defaults to two. Default: `0`.


### `resource.keep_alive.idle_connection_timeout` [_resource_keep_alive_idle_connection_timeout]

The maximum amount of time an idle connection will remain idle before closing itself. Valid time units are `ns`, `us`, `ms`, `s`, `m`, `h`. Zero means no limit. Default: `0s`.


### `resource.retry.max_attempts` [_resource_retry_max_attempts]

The maximum number of retries for the HTTP client. Default: `5`.


### `resource.retry.wait_min` [_resource_retry_wait_min]

The minimum time to wait before a retry is attempted. Default: `1s`.


### `resource.retry.wait_max` [_resource_retry_wait_max]

The maximum time to wait before a retry is attempted. Default: `60s`.


### `resource.redirect.forward_headers` [_resource_redirect_forward_headers]

When set to `true` request headers are forwarded in case of a redirect. Default: `false`.


### `resource.redirect.headers_ban_list` [_resource_redirect_headers_ban_list]

When `redirect.forward_headers` is set to `true`, all headers *except* the ones defined in this list will be forwarded. Default: `[]`.


### `resource.redirect.max_redirects` [_resource_redirect_max_redirects]

The maximum number of redirects to follow for a request. Default: `10`.


### `resource.max_body_size` [_resource_max_body_size]

```{applies_to}
stack: ga 8.18.1
```

The maximum size of a response body that will be accepted by the client if non-zero. Bodies that are too large will result in an error, "response body too big". Default: `0`.


### `resource.rate_limit.limit` [resource-rate-limit]

The value of the response that specifies the maximum overall resource request rate.


### `resource.rate_limit.burst` [_resource_rate_limit_burst]

The maximum burst size. Burst is the maximum number of resource requests that can be made above the overall rate limit.


### `resource.tracer.enable` [_resource_tracer_enable]

It is possible to log HTTP requests and responses in a CEL program to a local file-system for debugging configurations. This option is enabled by setting `resource.tracer.enabled` to true and setting the `resource.tracer.filename` value. Additional options are available to tune log rotation behavior. To delete existing logs, set `resource.tracer.enabled` to false without unsetting the filename option.

Enabling this option compromises security and should only be used for debugging. {applies_to}`stack: ga 8.15.0`


### `resource.tracer.filename` [_resource_tracer_filename]

To differentiate the trace files generated from different input instances, a placeholder `*` can be added to the filename and will be replaced with the input instance id. For Example, `http-request-trace-*.ndjson`.

Setting `resource.tracer.filename` with `resource.tracer.enable` set to false will cause any existing trace logs matching the filename option to be deleted. {applies_to}`stack: ga 8.15.0`


### `resource.tracer.maxsize` [_resource_tracer_maxsize]

This value sets the maximum size, in megabytes, the log file will reach before it is rotated. By default logs are allowed to reach 1MB before rotation. Individual request/response bodies will be truncated to 10% of this size.


### `resource.tracer.maxage` [_resource_tracer_maxage]

This specifies the number days to retain rotated log files. If it is not set, log files are retained indefinitely.


### `resource.tracer.maxbackups` [_resource_tracer_maxbackups]

The number of old logs to retain. If it is not set all old logs are retained subject to the `resource.tracer.maxage` setting.


### `resource.tracer.localtime` [_resource_tracer_localtime]

Whether to use the host’s local time rather that UTC for timestamping rotated log file names.


### `resource.tracer.compress` [_resource_tracer_compress]

This determines whether rotated logs should be gzip compressed.


### `redact` [cel-state-redact]

```{applies_to}
stack: ga 8.7.0
```

During debug level logging, the `state` object and the resulting evaluation result are included in logs. This may result in leaking of secrets. In order to prevent this, fields may be redacted or deleted from the logged `state`. The `redact` configuration allows users to configure this field redaction behavior. For safety reasons if the `redact` configuration is missing a warning is logged.

In the case of no-required redaction an empty `redact.fields` configuration should be used to silence the logged warning.

```yaml
- type: cel
  redact:
    fields: ~
```

As an example, if a user-constructed Basic Authentication request is used in a CEL program the password can be redacted like so

```yaml
filebeat.inputs:
- type: cel
  resource.url: http://localhost:9200/_search
  state:
    user: user@domain.tld
    password: P@$$W0₹D
  redact:
    fields:
      - password
    delete: true
```

Note that fields under the `auth` configuration hierarchy are not exposed to the `state` and so do not need to be redacted. For this reason it is preferable to use these for authentication over the request construction shown above where possible.


### `redact.fields` [_redact_fields]

This specifies fields in the `state` to be redacted prior to debug logging. Fields listed in this array will be either replaced with a `*` or deleted entirely from messages sent to debug logs.


### `redact.delete` [_redact_delete]

This specifies whether fields should be replaced with a `*` or deleted entirely from messages sent to debug logs. If delete is `true`, fields will be deleted rather than replaced.


### `failure_dump.enabled` [_failure_dump_enabled]

```{applies_to}
stack: ga 8.18.0
```

It is possible to log CEL program evaluation failures to a local file-system for debugging configurations. This option is enabled by setting `failure_dump.enabled` to true and setting the `failure_dump.filename` value. To delete existing failure dumps, set `failure_dump.enabled` to false without unsetting the filename option.

Enabling this option compromises security and should only be used for debugging.


### `failure_dump.filename` [_failure_dump_filename]

```{applies_to}
stack: ga 8.18.0
```

This specifies a directory path to write failure dumps to. If it is not empty and a CEL program evaluation fails, the complete set of states for the CEL program’s evaluation will be written as a JSON file, along with the error that was reported. This option should only be used when debugging a failure as it imposes a significant performance impact on the input and may potentially use large quantities of memory to hold the full set of states. If a failure dump is configured, it is recommended that data input sizes be reduced to avoid excessive memory consumption, and making dumps that are intractable to analysis. To delete existing failure dumps, set `failure_dump.enabled` to false without unsetting the filename option.


### `record_coverage` [cel-record-coverage]

```{applies_to}
stack: ga 8.18.0
```

This specifies that CEL code evaluation coverage should be recorded and logged in debug logs. This is a developer-only option.


## Metrics [_metrics_5]

This input exposes metrics under the [HTTP monitoring endpoint](/reference/filebeat/http-endpoint.md). These metrics are exposed under the `/inputs` path. They can be used to observe the activity of the input.

| Metric | Description |
| --- | --- |
| `resource` | URL or path of the input resource. |
| `cel_executions` | Number times the CEL program has been executed. |
| `batches_received_total` | Number of event arrays received. |
| `events_received_total` | Number of events received. |
| `batches_published_total` | Number of event arrays published. |
| `events_published_total` | Number of events published. |
| `cel_processing_time` | Histogram of the elapsed successful CEL program processing times in nanoseconds. |
| `batch_processing_time` | Histogram of the elapsed successful batch processing times in nanoseconds (time of receipt to time of ACK for non-empty batches). |
| `http_request_total` | Total number of processed requests. |
| `http_request_errors_total` | Total number of request errors. |
| `http_request_delete_total` | Total number of `DELETE` requests. |
| `http_request_get_total` | Total number of `GET` requests. |
| `http_request_head_total` | Total number of `HEAD` requests. |
| `http_request_options_total` | Total number of `OPTIONS` requests. |
| `http_request_patch_total` | Total number of `PATCH` requests. |
| `http_request_post_total` | Total number of `POST` requests. |
| `http_request_put_total` | Total number of `PUT` requests. |
| `http_request_body_bytes_total` | Total of the requests body size. |
| `http_request_body_bytes` | Histogram of the requests body size. |
| `http_response_total` | Total number of responses received. |
| `http_response_errors_total` | Total number of response errors. |
| `http_response_1xx_total` | Total number of `1xx` responses. |
| `http_response_2xx_total` | Total number of `2xx` responses. |
| `http_response_3xx_total` | Total number of `3xx` responses. |
| `http_response_4xx_total` | Total number of `4xx` responses. |
| `http_response_5xx_total` | Total number of `5xx` responses. |
| `http_response_body_bytes_total` | Total of the responses body size. |
| `http_response_body_bytes` | Histogram of the responses body size. |
| `http_round_trip_time` | Histogram of the round trip time. |


## Developer tools [_developer_tools]

A stand-alone CEL environment that implements the majority of the CEL input’s Comment Expression Language functionality is available in the [Elastic Mito](https://github.com/elastic/mito) repository. This tool may be used to help develop CEL programs to be used by the input. Installation is available from source by running `go install github.com/elastic/mito/cmd/mito@latest` and requires a Go toolchain.


## Common options [filebeat-input-cel-common-options]

The following configuration options are supported by all inputs.


#### `enabled` [_enabled_4]

Use the `enabled` option to enable and disable inputs. By default, enabled is set to true.


#### `tags` [_tags_4]

A list of tags that Filebeat includes in the `tags` field of each published event. Tags make it easy to select specific events in Kibana or apply conditional filtering in Logstash. These tags will be appended to the list of tags specified in the general configuration.

Example:

```yaml
filebeat.inputs:
- type: cel
  . . .
  tags: ["json"]
```


#### `fields` [filebeat-input-cel-fields]

Optional fields that you can specify to add additional information to the output. For example, you might add fields that you can use for filtering log data. Fields can be scalar values, arrays, dictionaries, or any nested combination of these. By default, the fields that you specify here will be grouped under a `fields` sub-dictionary in the output document. To store the custom fields as top-level fields, set the `fields_under_root` option to true. If a duplicate field is declared in the general configuration, then its value will be overwritten by the value declared here.

```yaml
filebeat.inputs:
- type: cel
  . . .
  fields:
    app_id: query_engine_12
```


#### `fields_under_root` [fields-under-root-cel]

If this option is set to true, the custom [fields](#filebeat-input-cel-fields) are stored as top-level fields in the output document instead of being grouped under a `fields` sub-dictionary. If the custom field names conflict with other field names added by Filebeat, then the custom fields overwrite the other fields.


#### `processors` [_processors_4]

A list of processors to apply to the input data.

See [Processors](/reference/filebeat/filtering-enhancing-data.md) for information about specifying processors in your config.


#### `pipeline` [_pipeline_4]

The ingest pipeline ID to set for the events generated by this input.

::::{note}
The pipeline ID can also be configured in the Elasticsearch output, but this option usually results in simpler configuration files. If the pipeline is configured both in the input and output, the option from the input is used.
::::


::::{important}
The `pipeline` is always lowercased. If `pipeline: Foo-Bar`, then the pipeline name in {{es}} needs to be defined as `foo-bar`.
::::



#### `keep_null` [_keep_null_4]

If this option is set to true, fields with `null` values will be published in the output document. By default, `keep_null` is set to `false`.


#### `index` [_index_4]

If present, this formatted string overrides the index for events from this input (for elasticsearch outputs), or sets the `raw_index` field of the event’s metadata (for other outputs). This string can only refer to the agent name and version and the event timestamp; for access to dynamic fields, use `output.elasticsearch.index` or a processor.

Example value: `"%{[agent.name]}-myindex-%{+yyyy.MM.dd}"` might expand to `"filebeat-myindex-2019.11.01"`.


#### `publisher_pipeline.disable_host` [_publisher_pipeline_disable_host_4]

By default, all events contain `host.name`. This option can be set to `true` to disable the addition of this field to all events. The default value is `false`.


