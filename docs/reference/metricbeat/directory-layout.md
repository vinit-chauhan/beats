---
mapped_pages:
  - https://www.elastic.co/guide/en/beats/metricbeat/current/directory-layout.html
applies_to:
  stack: ga
---

# Directory layout [directory-layout]

The directory layout of an installation is as follows:

::::{tip}
Archive installation has a different layout. See [zip, tar.gz, or tgz](#directory-layout-archive).
::::


| Type | Description | Default Location | Config Option |
| --- | --- | --- | --- |
| home | Home of the Metricbeat installation. |  | `path.home` |
| bin | The location for the binary files. | `{path.home}/bin` |  |
| config | The location for configuration files. | `{path.home}` | `path.config` |
| data | The location for persistent data files. | `{path.home}/data` | `path.data` |
| logs | The location for the logs created by Metricbeat. | `{path.home}/logs` | `path.logs` |

You can change these settings by using CLI flags or setting [path options](/reference/metricbeat/configuration-path.md) in the configuration file.

## Default paths [_default_paths]

Metricbeat uses the following default paths unless you explicitly change them.


#### deb and rpm [_deb_and_rpm]

| Type | Description | Location |
| --- | --- | --- |
| home | Home of the Metricbeat installation. | `/usr/share/metricbeat` |
| bin | The location for the binary files. | `/usr/share/metricbeat/bin` |
| config | The location for configuration files. | `/etc/metricbeat` |
| data | The location for persistent data files. | `/var/lib/metricbeat` |
| logs | The location for the logs created by Metricbeat. | `/var/log/metricbeat` |

For the deb and rpm distributions, these paths are set in the init script or in the systemd unit file.  Make sure that you start the Metricbeat service by using the preferred operating system method (init scripts or `systemctl`). Otherwise the paths might be set incorrectly.


#### docker [_docker]

| Type | Description | Location |
| --- | --- | --- |
| home | Home of the Metricbeat installation. | `/usr/share/metricbeat` |
| bin | The location for the binary files. | `/usr/share/metricbeat` |
| config | The location for configuration files. | `/usr/share/metricbeat` |
| data | The location for persistent data files. | `/usr/share/metricbeat/data` |
| logs | The location for the logs created by Metricbeat. | `/usr/share/metricbeat/logs` |


#### zip, tar.gz, or tgz [directory-layout-archive]

| Type | Description | Location |
| --- | --- | --- |
| home | Home of the Metricbeat installation. | `{extract.path}` |
| bin | The location for the binary files. | `{extract.path}` |
| config | The location for configuration files. | `{extract.path}` |
| data | The location for persistent data files. | `{extract.path}/data` |
| logs | The location for the logs created by Metricbeat. | `{extract.path}/logs` |

For the zip, tar.gz, or tgz distributions, these paths are based on the location of the extracted binary file. This means that if you start Metricbeat with the following simple command, all paths are set correctly:

```sh
./metricbeat
```


