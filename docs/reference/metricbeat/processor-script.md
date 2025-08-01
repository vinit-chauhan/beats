---
navigation_title: "script"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/metricbeat/current/processor-script.html
applies_to:
  stack: ga
---

# Script Processor [processor-script]


The `script` processor executes Javascript code to process an event. The processor uses a pure Go implementation of ECMAScript 5.1 and has no external dependencies. This can be useful in situations where one of the other processors doesn’t provide the functionality you need to filter events.

The processor can be configured by embedding Javascript in your configuration file or by pointing the processor at external file(s).

```yaml
processors:
  - script:
      lang: javascript
      source: >
        function process(event) {
            event.Tag("js");
        }
```

This loads `filter.js` from disk.

```yaml
processors:
  - script:
      lang: javascript
      file: ${path.config}/filter.js
```

Parameters can be passed to the script by adding `params` to the config. This allows for a script to be made reusable. When using `params` the code must define a `register(params)` function to receive the parameters.

```yaml
processors:
  - script:
      lang: javascript
      tag: my_filter
      params:
        threshold: 15
      source: >
        var params = {threshold: 42};
        function register(scriptParams) {
            params = scriptParams;
        }
        function process(event) {
            if (event.Get("severity") < params.threshold) {
                event.Cancel();
            }
        }
```

If the script defines a `test()` function it will be invoked when the processor is loaded. Any exceptions thrown will cause the processor to fail to load. This can be used to make assertions about the behavior of the script.

```javascript
function process(event) {
    if (event.Get("event.code") === 1102) {
        event.Put("event.action", "cleared");
    }
    return event;
}

function test() {
    var event = process(new Event({event: {code: 1102}}));
    if (event.Get("event.action") !== "cleared") {
        throw "expected event.action === cleared";
    }
}
```


## Configuration options [_configuration_options_13]

The `script` processor has the following configuration settings:

`lang`
:   This field is required and its value must be `javascript`.

`tag`
:   This is an optional identifier that is added to log messages. If defined it enables metrics logging for this instance of the processor. The metrics include the number of exceptions and a histogram of the execution times for the `process` function.

`source`
:   Inline Javascript source code.

`file`
:   Path to a script file to load. Relative paths are interpreted as relative to the `path.config` directory. Globs are expanded.

`files`
:   List of script files to load. The scripts are concatenated together. Relative paths are interpreted as relative to the `path.config` directory. And globs are expanded.

`params`
:   A dictionary of parameters that are passed to the `register` of the script.

`tag_on_exception`
:   Tag to add to events in case the Javascript code causes an exception while processing an event. Defaults to `_js_exception`.

`timeout`
:   This sets an execution timeout for the `process` function. When the `process` function takes longer than the `timeout` period the function is interrupted. You can set this option to prevent a script from running for too long (like preventing an infinite `while` loop). By default there is no timeout.

`max_cached_sessions`
:   This sets the maximum number of Javascript VM sessions that will be cached to avoid reallocation. The default is `4`.


## Event API [_event_api]

The `Event` object passed to the `process` method has the following API.

| Method | Description |
| --- | --- |
| `Get(string)` | Get a value from the event (either a scalar or an object). If the key does notexist `null` is returned. If no key is provided then an object containing allfields is returned.<br>**Example**: `var value = event.Get(key);` |
| `Put(string, value)` | Put a value into the event. If the key was already set then theprevious value is returned. It throws an exception if the key cannot be setbecause one of the intermediate values is not an object.<br>**Example**: `var old = event.Put(key, value);` |
| `Rename(string, string)` | Rename a key in the event. The target key must not exist. Itreturns true if the source key was successfully renamed to the target key.<br>**Example**: `var success = event.Rename("source", "target");` |
| `Delete(string)` | Delete a field from the event. It returns true on success.<br>**Example**: `var deleted = event.Delete("user.email");` |
| `Cancel()` | Flag the event as cancelled which causes the processor to dropevent.<br>**Example**: `event.Cancel(); return;` |
| `Tag(string)` | Append a tag to the `tags` field if the tag does not alreadyexist. Throws an exception if `tags` exists and is not a string or a list ofstrings.<br>**Example**: `event.Tag("user_event");` |
| `AppendTo(string, string)` | `AppendTo` is a specialized `Put` method that converts the existing value to anarray and appends the value if it does not already exist. If there is anexisting value that’s not a string or array of strings then an exception isthrown.<br>**Example**: `event.AppendTo("error.message", "invalid file hash");` |

