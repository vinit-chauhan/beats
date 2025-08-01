---
navigation_title: "SSL"
mapped_pages:
  - https://www.elastic.co/guide/en/beats/metricbeat/current/configuration-ssl.html
applies_to:
  stack: ga
---

# Configure SSL [configuration-ssl]


You can specify SSL options when you configure:

* [outputs](/reference/metricbeat/configuring-output.md) that support SSL
* the [Kibana endpoint](/reference/metricbeat/setup-kibana-endpoint.md)
* [modules](/reference/metricbeat/metricbeat-modules.md) that define the host as an HTTP URL

Example output config with SSL enabled:

```yaml
output.elasticsearch.hosts: ["https://192.168.1.42:9200"]
output.elasticsearch.ssl.certificate_authorities: ["/etc/client/ca.pem"]
output.elasticsearch.ssl.certificate: "/etc/client/cert.pem"
output.elasticsearch.ssl.key: "/etc/client/cert.key"
```

Also see [*Secure communication with Logstash*](/reference/metricbeat/configuring-ssl-logstash.md).

Example Kibana endpoint config with SSL enabled:

```yaml
setup.kibana.host: "https://192.0.2.255:5601"
setup.kibana.ssl.enabled: true
setup.kibana.ssl.certificate_authorities: ["/etc/client/ca.pem"]
setup.kibana.ssl.certificate: "/etc/client/cert.pem"
setup.kibana.ssl.key: "/etc/client/cert.key"
```

Example module with SSL enabled:

```yaml
- module: http
  namespace: "myservice"
  enabled: true
  period: 10s
  hosts: ["https://localhost"]
  path: "/stats"
  headers:
    Authorization: "Bearer test123"
  ssl.verification_mode: "none"
```

There are a number of SSL configuration options available to you:

* [Common configuration options](#ssl-common-config)
* [Client configuration options](#ssl-client-config)
* [Server configuration options](#ssl-server-config)


## Common configuration options [ssl-common-config]

Common SSL configuration options can be used in both client and server configurations. You can specify the following options in the `ssl` section of each subsystem that supports SSL.


### `enabled` [enabled]

To disable SSL configuration, set the value to `false`. The default value is `true`.

::::{note}
SSL settings are disabled if either `enabled` is set to `false` or the `ssl` section is missing.

::::



### `supported_protocols` [supported-protocols]

List of allowed SSL/TLS versions. If SSL/TLS server decides for protocol versions not configured, the connection will be dropped during or after the handshake. The setting is a list of allowed protocol versions: `TLSv1.1`, `TLSv1.2`, and `TLSv1.3`.

The default value is `[TLSv1.2, TLSv1.3]`.


### `cipher_suites` [cipher-suites]

The list of cipher suites to use. The first entry has the highest priority. If this option is omitted, the Go crypto library’s [default suites](https://golang.org/pkg/crypto/tls/) are used (recommended).

Note that if TLS 1.3 is enabled (which is true by default), then the default TLS 1.3 cipher suites are always included, because Go’s standard library adds them to all connections. In order to exclude the default TLS 1.3 ciphers, TLS 1.3 must also be disabled, e.g. with the setting `ssl.supported_protocols = [TLSv1.2]`.

The following cipher suites are available:

| Cypher | Notes |
| --- | --- |
| ECDHE-ECDSA-AES-128-CBC-SHA |  |
| ECDHE-ECDSA-AES-128-CBC-SHA256 | TLS 1.2 only. Disabled by default. |
| ECDHE-ECDSA-AES-128-GCM-SHA256 | TLS 1.2 only. |
| ECDHE-ECDSA-AES-256-CBC-SHA |  |
| ECDHE-ECDSA-AES-256-GCM-SHA384 | TLS 1.2 only. |
| ECDHE-ECDSA-CHACHA20-POLY1305 | TLS 1.2 only. |
| ECDHE-ECDSA-RC4-128-SHA | Disabled by default. RC4 not recommended. |
| ECDHE-RSA-3DES-CBC3-SHA |  |
| ECDHE-RSA-AES-128-CBC-SHA |  |
| ECDHE-RSA-AES-128-CBC-SHA256 | TLS 1.2 only. Disabled by default. |
| ECDHE-RSA-AES-128-GCM-SHA256 | TLS 1.2 only. |
| ECDHE-RSA-AES-256-CBC-SHA |  |
| ECDHE-RSA-AES-256-GCM-SHA384 | TLS 1.2 only. |
| ECDHE-RSA-CHACHA20-POLY1205 | TLS 1.2 only. |
| ECDHE-RSA-RC4-128-SHA | Disabled by default. RC4 not recommended. |
| RSA-3DES-CBC3-SHA |  |
| RSA-AES-128-CBC-SHA |  |
| RSA-AES-128-CBC-SHA256 | TLS 1.2 only. Disabled by default. |
| RSA-AES-128-GCM-SHA256 | TLS 1.2 only. |
| RSA-AES-256-CBC-SHA |  |
| RSA-AES-256-GCM-SHA384 | TLS 1.2 only. |
| RSA-RC4-128-SHA | Disabled by default. RC4 not recommended. |

Here is a list of acronyms used in defining the cipher suites:

* 3DES: Cipher suites using triple DES
* AES-128/256: Cipher suites using AES with 128/256-bit keys.
* CBC: Cipher using Cipher Block Chaining as block cipher mode.
* ECDHE: Cipher suites using Elliptic Curve Diffie-Hellman (DH) ephemeral key exchange.
* ECDSA: Cipher suites using Elliptic Curve Digital Signature Algorithm for authentication.
* GCM: Galois/Counter mode is used for symmetric key cryptography.
* RC4: Cipher suites using RC4.
* RSA: Cipher suites using RSA.
* SHA, SHA256, SHA384: Cipher suites using SHA-1, SHA-256 or SHA-384.


### `curve_types` [curve-types]

The list of curve types for ECDHE (Elliptic Curve Diffie-Hellman ephemeral key exchange).

The following elliptic curve types are available:

* P-256
* P-384
* P-521
* X25519


### `ca_sha256` [ca-sha256]

This configures a certificate pin that you can use to ensure that a specific certificate is part of the verified chain.

The pin is a base64 encoded string of the SHA-256 of the certificate.

::::{note}
This check is not a replacement for the normal SSL validation, but it adds additional validation. If this option is used with  `verification_mode` set to `none`, the check will always fail because it will not receive any verified chains.
::::



## Client configuration options [ssl-client-config]

You can specify the following options in the `ssl` section of each subsystem that supports SSL.


### `certificate_authorities` [client-certificate-authorities]

The list of root certificates for verifications is required. If `certificate_authorities` is empty or not set, the system keystore is used. If `certificate_authorities` is self-signed, the host system needs to trust that CA cert as well.

By default you can specify a list of files that `metricbeat` will read, but you can also embed a certificate directly in the `YAML` configuration:

```yaml
certificate_authorities:
  - |
    -----BEGIN CERTIFICATE-----
    MIIDCjCCAfKgAwIBAgITJ706Mu2wJlKckpIvkWxEHvEyijANBgkqhkiG9w0BAQsF
    ADAUMRIwEAYDVQQDDAlsb2NhbGhvc3QwIBcNMTkwNzIyMTkyOTA0WhgPMjExOTA2
    MjgxOTI5MDRaMBQxEjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEB
    BQADggEPADCCAQoCggEBANce58Y/JykI58iyOXpxGfw0/gMvF0hUQAcUrSMxEO6n
    fZRA49b4OV4SwWmA3395uL2eB2NB8y8qdQ9muXUdPBWE4l9rMZ6gmfu90N5B5uEl
    94NcfBfYOKi1fJQ9i7WKhTjlRkMCgBkWPkUokvBZFRt8RtF7zI77BSEorHGQCk9t
    /D7BS0GJyfVEhftbWcFEAG3VRcoMhF7kUzYwp+qESoriFRYLeDWv68ZOvG7eoWnP
    PsvZStEVEimjvK5NSESEQa9xWyJOmlOKXhkdymtcUd/nXnx6UTCFgnkgzSdTWV41
    CI6B6aJ9svCTI2QuoIq2HxX/ix7OvW1huVmcyHVxyUECAwEAAaNTMFEwHQYDVR0O
    BBYEFPwN1OceFGm9v6ux8G+DZ3TUDYxqMB8GA1UdIwQYMBaAFPwN1OceFGm9v6ux
    8G+DZ3TUDYxqMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAG5D
    874A4YI7YUwOVsVAdbWtgp1d0zKcPRR+r2OdSbTAV5/gcS3jgBJ3i1BN34JuDVFw
    3DeJSYT3nxy2Y56lLnxDeF8CUTUtVQx3CuGkRg1ouGAHpO/6OqOhwLLorEmxi7tA
    H2O8mtT0poX5AnOAhzVy7QW0D/k4WaoLyckM5hUa6RtvgvLxOwA0U+VGurCDoctu
    8F4QOgTAWyh8EZIwaKCliFRSynDpv3JTUwtfZkxo6K6nce1RhCWFAsMvDZL8Dgc0
    yvgJ38BRsFOtkRuAGSf6ZUwTO8JJRRIFnpUzXflAnGivK9M13D5GEQMmIl6U9Pvk
    sxSmbIUfc2SGJGCJD4I=
    -----END CERTIFICATE-----
```


### `certificate: "/etc/client/cert.pem"` [client-certificate]

The path to the certificate for SSL client authentication is only required if `client_authentication` is specified. If the certificate is not specified, client authentication is not available. The connection might fail if the server requests client authentication. If the SSL server does not require client authentication, the certificate will be loaded, but not requested or used by the server.

When this option is configured, the [`key`](#client-key) option is also required. The certificate option support embedding of the certificate:

```yaml
certificate: |
    -----BEGIN CERTIFICATE-----
    MIIDCjCCAfKgAwIBAgITJ706Mu2wJlKckpIvkWxEHvEyijANBgkqhkiG9w0BAQsF
    ADAUMRIwEAYDVQQDDAlsb2NhbGhvc3QwIBcNMTkwNzIyMTkyOTA0WhgPMjExOTA2
    MjgxOTI5MDRaMBQxEjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEB
    BQADggEPADCCAQoCggEBANce58Y/JykI58iyOXpxGfw0/gMvF0hUQAcUrSMxEO6n
    fZRA49b4OV4SwWmA3395uL2eB2NB8y8qdQ9muXUdPBWE4l9rMZ6gmfu90N5B5uEl
    94NcfBfYOKi1fJQ9i7WKhTjlRkMCgBkWPkUokvBZFRt8RtF7zI77BSEorHGQCk9t
    /D7BS0GJyfVEhftbWcFEAG3VRcoMhF7kUzYwp+qESoriFRYLeDWv68ZOvG7eoWnP
    PsvZStEVEimjvK5NSESEQa9xWyJOmlOKXhkdymtcUd/nXnx6UTCFgnkgzSdTWV41
    CI6B6aJ9svCTI2QuoIq2HxX/ix7OvW1huVmcyHVxyUECAwEAAaNTMFEwHQYDVR0O
    BBYEFPwN1OceFGm9v6ux8G+DZ3TUDYxqMB8GA1UdIwQYMBaAFPwN1OceFGm9v6ux
    8G+DZ3TUDYxqMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAG5D
    874A4YI7YUwOVsVAdbWtgp1d0zKcPRR+r2OdSbTAV5/gcS3jgBJ3i1BN34JuDVFw
    3DeJSYT3nxy2Y56lLnxDeF8CUTUtVQx3CuGkRg1ouGAHpO/6OqOhwLLorEmxi7tA
    H2O8mtT0poX5AnOAhzVy7QW0D/k4WaoLyckM5hUa6RtvgvLxOwA0U+VGurCDoctu
    8F4QOgTAWyh8EZIwaKCliFRSynDpv3JTUwtfZkxo6K6nce1RhCWFAsMvDZL8Dgc0
    yvgJ38BRsFOtkRuAGSf6ZUwTO8JJRRIFnpUzXflAnGivK9M13D5GEQMmIl6U9Pvk
    sxSmbIUfc2SGJGCJD4I=
    -----END CERTIFICATE-----
```


### `key: "/etc/client/cert.key"` [client-key]

The client certificate key used for client authentication and is only required if `client_authentication` is configured. The key option support embedding of the private key:

```yaml
key: |
    -----BEGIN PRIVATE KEY-----
    MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDXHufGPycpCOfI
    sjl6cRn8NP4DLxdIVEAHFK0jMRDup32UQOPW+DleEsFpgN9/ebi9ngdjQfMvKnUP
    Zrl1HTwVhOJfazGeoJn7vdDeQebhJfeDXHwX2DiotXyUPYu1ioU45UZDAoAZFj5F
    KJLwWRUbfEbRe8yO+wUhKKxxkApPbfw+wUtBicn1RIX7W1nBRABt1UXKDIRe5FM2
    MKfqhEqK4hUWC3g1r+vGTrxu3qFpzz7L2UrRFRIpo7yuTUhEhEGvcVsiTppTil4Z
    HcprXFHf5158elEwhYJ5IM0nU1leNQiOgemifbLwkyNkLqCKth8V/4sezr1tYblZ
    nMh1cclBAgMBAAECggEBAKdP5jyOicqknoG9/G564RcDsDyRt64NuO7I6hBg7SZx
    Jn7UKWDdFuFP/RYtoabn6QOxkVVlydp5Typ3Xu7zmfOyss479Q/HIXxmmbkD0Kp0
    eRm2KN3y0b6FySsS40KDRjKGQCuGGlNotW3crMw6vOvvsLTlcKgUHF054UVCHoK/
    Piz7igkDU7NjvJeha53vXL4hIjb10UtJNaGPxIyFLYRZdRPyyBJX7Yt3w8dgz8WM
    epOPu0dq3bUrY3WQXcxKZo6sQjE1h7kdl4TNji5jaFlvD01Y8LnyG0oThOzf0tve
    Gaw+kuy17gTGZGMIfGVcdeb+SlioXMAAfOps+mNIwTECgYEA/gTO8W0hgYpOQJzn
    BpWkic3LAoBXWNpvsQkkC3uba8Fcps7iiEzotXGfwYcb5Ewf5O3Lrz1EwLj7GTW8
    VNhB3gb7bGOvuwI/6vYk2/dwo84bwW9qRWP5hqPhNZ2AWl8kxmZgHns6WTTxpkRU
    zrfZ5eUrBDWjRU2R8uppgRImsxMCgYEA2MxuL/C/Ko0d7XsSX1kM4JHJiGpQDvb5
    GUrlKjP/qVyUysNF92B9xAZZHxxfPWpdfGGBynhw7X6s+YeIoxTzFPZVV9hlkpAA
    5igma0n8ZpZEqzttjVdpOQZK8o/Oni/Q2S10WGftQOOGw5Is8+LY30XnLvHBJhO7
    TKMurJ4KCNsCgYAe5TDSVmaj3dGEtFC5EUxQ4nHVnQyCpxa8npL+vor5wSvmsfUF
    hO0s3GQE4sz2qHecnXuPldEd66HGwC1m2GKygYDk/v7prO1fQ47aHi9aDQB9N3Li
    e7Vmtdn3bm+lDjtn0h3Qt0YygWj+wwLZnazn9EaWHXv9OuEMfYxVgYKpdwKBgEze
    Zy8+WDm5IWRjn8cI5wT1DBT/RPWZYgcyxABrwXmGZwdhp3wnzU/kxFLAl5BKF22T
    kRZ+D+RVZvVutebE9c937BiilJkb0AXLNJwT9pdVLnHcN2LHHHronUhV7vetkop+
    kGMMLlY0lkLfoGq1AxpfSbIea9KZam6o6VKxEnPDAoGAFDCJm+ZtsJK9nE5GEMav
    NHy+PwkYsHhbrPl4dgStTNXLenJLIJ+Ke0Pcld4ZPfYdSyu/Tv4rNswZBNpNsW9K
    0NwJlyMBfayoPNcJKXrH/csJY7hbKviAHr1eYy9/8OL0dHf85FV+9uY5YndLcsDc
    nygO9KTJuUiBrLr0AHEnqko=
    -----END PRIVATE KEY-----
```


### `key_passphrase` [client-key-passphrase]

The passphrase used to decrypt an encrypted key stored in the configured `key` file.


### `verification_mode` [client-verification-mode]

Controls the verification of server certificates. Valid values are:

`full`
:   Verifies that the provided certificate is signed by a trusted authority (CA) and also verifies that the server’s hostname (or IP address) matches the names identified within the certificate.

`strict`
:   Verifies that the provided certificate is signed by a trusted authority (CA) and also verifies that the server’s hostname (or IP address) matches the names identified within the certificate. If the Subject Alternative Name is empty, it returns an error.

`certificate`
:   Verifies that the provided certificate is signed by a trusted authority (CA), but does not perform any hostname verification.

`none`
:   Performs *no verification* of the server’s certificate. This mode disables many of the security benefits of SSL/TLS and should only be used after cautious consideration. It is primarily intended as a temporary diagnostic mechanism when attempting to resolve TLS errors; its use in production environments is strongly discouraged.

    The default value is `full`.



### `ca_trusted_fingerprint` [ca_trusted_fingerprint]

A HEX encoded SHA-256 of a CA certificate. If this certificate is present in the chain during the handshake, it will be added to the `certificate_authorities` list and the handshake will continue normaly.

To get the fingerprint from a CA certificate on a Unix-like system, you can use the following command, where `ca.crt` is the certificate.

```
openssl x509 -fingerprint -sha256 -noout -in ./ca.crt | awk --field-separator="=" '{print $2}' | sed 's/://g'
```


## Server configuration options [ssl-server-config]

You can specify the following options in the `ssl` section of each subsystem that supports SSL.


### `certificate_authorities` [server-certificate-authorities]

The list of root certificates for client verifications is only required if `client_authentication` is configured. If `certificate_authorities` is empty or not set, and `client_authentication` is configured, the system keystore is used.

If `certificate_authorities` is self-signed, the host system needs to trust that CA cert as well. By default you can specify a list of files that `metricbeat` will read, but you can also embed a certificate directly in the `YAML` configuration:

```yaml
certificate_authorities:
  - |
    -----BEGIN CERTIFICATE-----
    MIIDCjCCAfKgAwIBAgITJ706Mu2wJlKckpIvkWxEHvEyijANBgkqhkiG9w0BAQsF
    ADAUMRIwEAYDVQQDDAlsb2NhbGhvc3QwIBcNMTkwNzIyMTkyOTA0WhgPMjExOTA2
    MjgxOTI5MDRaMBQxEjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEB
    BQADggEPADCCAQoCggEBANce58Y/JykI58iyOXpxGfw0/gMvF0hUQAcUrSMxEO6n
    fZRA49b4OV4SwWmA3395uL2eB2NB8y8qdQ9muXUdPBWE4l9rMZ6gmfu90N5B5uEl
    94NcfBfYOKi1fJQ9i7WKhTjlRkMCgBkWPkUokvBZFRt8RtF7zI77BSEorHGQCk9t
    /D7BS0GJyfVEhftbWcFEAG3VRcoMhF7kUzYwp+qESoriFRYLeDWv68ZOvG7eoWnP
    PsvZStEVEimjvK5NSESEQa9xWyJOmlOKXhkdymtcUd/nXnx6UTCFgnkgzSdTWV41
    CI6B6aJ9svCTI2QuoIq2HxX/ix7OvW1huVmcyHVxyUECAwEAAaNTMFEwHQYDVR0O
    BBYEFPwN1OceFGm9v6ux8G+DZ3TUDYxqMB8GA1UdIwQYMBaAFPwN1OceFGm9v6ux
    8G+DZ3TUDYxqMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAG5D
    874A4YI7YUwOVsVAdbWtgp1d0zKcPRR+r2OdSbTAV5/gcS3jgBJ3i1BN34JuDVFw
    3DeJSYT3nxy2Y56lLnxDeF8CUTUtVQx3CuGkRg1ouGAHpO/6OqOhwLLorEmxi7tA
    H2O8mtT0poX5AnOAhzVy7QW0D/k4WaoLyckM5hUa6RtvgvLxOwA0U+VGurCDoctu
    8F4QOgTAWyh8EZIwaKCliFRSynDpv3JTUwtfZkxo6K6nce1RhCWFAsMvDZL8Dgc0
    yvgJ38BRsFOtkRuAGSf6ZUwTO8JJRRIFnpUzXflAnGivK9M13D5GEQMmIl6U9Pvk
    sxSmbIUfc2SGJGCJD4I=
    -----END CERTIFICATE-----
```


### `certificate: "/etc/server/cert.pem"` [server-certificate]

The end-entity (leaf) certificate that the server uses to identify itself. If the certificate is signed by a certificate authority (CA), then it should include intermediate CA certificates, sorted from leaf to root. For servers, a `certificate` and [`key`](#server-key) must be specified.

The certificate option supports embedding of the PEM certificate content. This example contains the leaf certificate followed by issuer’s certificate.

```yaml
certificate: |
  -----BEGIN CERTIFICATE-----
  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
  puoJ9UYvfIcHUGPdZGU=
  -----END CERTIFICATE-----
  -----BEGIN CERTIFICATE-----
  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
  RQQmEyTn1hn/eA==
  -----END CERTIFICATE-----
```


### `key: "/etc/server/cert.key"` [server-key]

The server certificate key used for authentication is required. The key option supports embedding of the private key:

```yaml
key: |
    -----BEGIN PRIVATE KEY-----
    MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDXHufGPycpCOfI
    sjl6cRn8NP4DLxdIVEAHFK0jMRDup32UQOPW+DleEsFpgN9/ebi9ngdjQfMvKnUP
    Zrl1HTwVhOJfazGeoJn7vdDeQebhJfeDXHwX2DiotXyUPYu1ioU45UZDAoAZFj5F
    KJLwWRUbfEbRe8yO+wUhKKxxkApPbfw+wUtBicn1RIX7W1nBRABt1UXKDIRe5FM2
    MKfqhEqK4hUWC3g1r+vGTrxu3qFpzz7L2UrRFRIpo7yuTUhEhEGvcVsiTppTil4Z
    HcprXFHf5158elEwhYJ5IM0nU1leNQiOgemifbLwkyNkLqCKth8V/4sezr1tYblZ
    nMh1cclBAgMBAAECggEBAKdP5jyOicqknoG9/G564RcDsDyRt64NuO7I6hBg7SZx
    Jn7UKWDdFuFP/RYtoabn6QOxkVVlydp5Typ3Xu7zmfOyss479Q/HIXxmmbkD0Kp0
    eRm2KN3y0b6FySsS40KDRjKGQCuGGlNotW3crMw6vOvvsLTlcKgUHF054UVCHoK/
    Piz7igkDU7NjvJeha53vXL4hIjb10UtJNaGPxIyFLYRZdRPyyBJX7Yt3w8dgz8WM
    epOPu0dq3bUrY3WQXcxKZo6sQjE1h7kdl4TNji5jaFlvD01Y8LnyG0oThOzf0tve
    Gaw+kuy17gTGZGMIfGVcdeb+SlioXMAAfOps+mNIwTECgYEA/gTO8W0hgYpOQJzn
    BpWkic3LAoBXWNpvsQkkC3uba8Fcps7iiEzotXGfwYcb5Ewf5O3Lrz1EwLj7GTW8
    VNhB3gb7bGOvuwI/6vYk2/dwo84bwW9qRWP5hqPhNZ2AWl8kxmZgHns6WTTxpkRU
    zrfZ5eUrBDWjRU2R8uppgRImsxMCgYEA2MxuL/C/Ko0d7XsSX1kM4JHJiGpQDvb5
    GUrlKjP/qVyUysNF92B9xAZZHxxfPWpdfGGBynhw7X6s+YeIoxTzFPZVV9hlkpAA
    5igma0n8ZpZEqzttjVdpOQZK8o/Oni/Q2S10WGftQOOGw5Is8+LY30XnLvHBJhO7
    TKMurJ4KCNsCgYAe5TDSVmaj3dGEtFC5EUxQ4nHVnQyCpxa8npL+vor5wSvmsfUF
    hO0s3GQE4sz2qHecnXuPldEd66HGwC1m2GKygYDk/v7prO1fQ47aHi9aDQB9N3Li
    e7Vmtdn3bm+lDjtn0h3Qt0YygWj+wwLZnazn9EaWHXv9OuEMfYxVgYKpdwKBgEze
    Zy8+WDm5IWRjn8cI5wT1DBT/RPWZYgcyxABrwXmGZwdhp3wnzU/kxFLAl5BKF22T
    kRZ+D+RVZvVutebE9c937BiilJkb0AXLNJwT9pdVLnHcN2LHHHronUhV7vetkop+
    kGMMLlY0lkLfoGq1AxpfSbIea9KZam6o6VKxEnPDAoGAFDCJm+ZtsJK9nE5GEMav
    NHy+PwkYsHhbrPl4dgStTNXLenJLIJ+Ke0Pcld4ZPfYdSyu/Tv4rNswZBNpNsW9K
    0NwJlyMBfayoPNcJKXrH/csJY7hbKviAHr1eYy9/8OL0dHf85FV+9uY5YndLcsDc
    nygO9KTJuUiBrLr0AHEnqko=
    -----END PRIVATE KEY-----
```


### `key_passphrase` [server-key-passphrase]

The passphrase is used to decrypt an encrypted key stored in the configured `key` file.


### `verification_mode` [server-verification-mode]

Controls the verification of client certificates. Valid values are:

`full`
:   Verifies that the provided certificate is signed by a trusted authority (CA) and also verifies that the server’s hostname (or IP address) matches the names identified within the certificate.

`strict`
:   Verifies that the provided certificate is signed by a trusted authority (CA) and also verifies that the server’s hostname (or IP address) matches the names identified within the certificate. If the Subject Alternative Name is empty, it returns an error.

`certificate`
:   Verifies that the provided certificate is signed by a trusted authority (CA), but does not perform any hostname verification.

`none`
:   Performs *no verification* of the server’s certificate. This mode disables many of the security benefits of SSL/TLS and should only be used after cautious consideration. It is primarily intended as a temporary diagnostic mechanism when attempting to resolve TLS errors; its use in production environments is strongly discouraged.

    The default value is `full`.



### `renegotiation` [server-renegotiation]

This configures what types of TLS renegotiation are supported. The valid options are:

`never`
:   Disables renegotiation.

`once`
:   Allows a remote server to request renegotiation once per connection.

`freely`
:   Allows a remote server to request renegotiation repeatedly.

    The default value is `never`.



### `restart_on_cert_change.enabled` [exit_on_cert_change_enabled]

If set to `true` Metricbeat will restart if any file listed by `key`, `certificate`, or `certificate_authorities` is modified.

::::{note}
This feature is NOT supported on Windows. The default value is `false`.
::::


::::{note}
This feature requres the `execve` system call to be enabled. If you have a custom seccomp policy in place, make sure to allow for `execve`.
::::



### `restart_on_cert_change.period` [restart_on_cert_change_period]

Specifies how often the files are checked for changes. Do not set the period to less than 1s because the modification time of files is often stored in seconds. Setting the period to less than 1s will result in validation error and Metricbeat will not start. The default value is 1m.

