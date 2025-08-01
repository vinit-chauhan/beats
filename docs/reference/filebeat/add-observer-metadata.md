---
navigation_title: "add_observer_metadata"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/filebeat/current/add-observer-metadata.html
applies_to:
  stack: beta
---

# Add Observer metadata [add-observer-metadata]


::::{warning}
This functionality is in beta and is subject to change. The design and code is less mature than official GA features and is being provided as-is with no warranties. Beta features are not subject to the support SLA of official GA features.
::::


```yaml
processors:
  - add_observer_metadata:
      cache.ttl: 5m
      geo:
        name: nyc-dc1-rack1
        location: 40.7128, -74.0060
        continent_name: North America
        country_iso_code: US
        region_name: New York
        region_iso_code: NY
        city_name: New York
```

It has the following settings:

`netinfo.enabled`
:   (Optional) Default true. Include IP addresses and MAC addresses as fields observer.ip and observer.mac

`cache.ttl`
:   (Optional) The processor uses an internal cache for the observer metadata. This sets the cache expiration time. The default is 5m, negative values disable caching altogether.

`geo.name`
:   (Optional) User definable token to be used for identifying a discrete location. Frequently a datacenter, rack, or similar.

`geo.location`
:   (Optional) Longitude and latitude in comma separated format.

`geo.continent_name`
:   (Optional) Name of the continent.

`geo.country_name`
:   (Optional) Name of the country.

`geo.region_name`
:   (Optional) Name of the region.

`geo.city_name`
:   (Optional) Name of the city.

`geo.country_iso_code`
:   (Optional) ISO country code.

`geo.region_iso_code`
:   (Optional) ISO region code.

The `add_observer_metadata` processor annotates each event with relevant metadata from the observer machine. The fields added to the event look like the following:

```json
{
  "observer" : {
    "hostname" : "avce",
    "type" : "heartbeat",
    "vendor" : "elastic",
    "ip" : [
      "192.168.1.251",
      "fe80::64b2:c3ff:fe5b:b974",
    ],
    "mac" : [
      "dc:c1:02:6f:1b:ed",
    ],
    "geo": {
      "continent_name": "North America",
      "country_iso_code": "US",
      "region_name": "New York",
      "region_iso_code": "NY",
      "city_name": "New York",
      "name": "nyc-dc1-rack1",
      "location": "40.7128, -74.0060"
    }
  }
}
```

