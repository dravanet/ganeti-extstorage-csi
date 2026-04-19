# ganeti-extstorage-csi

Ganeti support for CSI through external storage interface.

## Architecture

```ascii
+--------------------+     +-----------------------+     +------------+
|                    |     |                       |     |            |
|       Ganeti       +-----> ganeti-extstorage-csi +-----> csi-driver |
|                    |     |                       |     |            |
+--------------------+     +------------+----------+     +------------+
                                        |
                                        |
                                        |
                                        |
                              +---------v----------+
                              |                    |
                              |  Metadata storage  |
                              |                    |
                              +--------------------+
```

ganeti-extstorage-csi implements [ganeti-extstorage-interface](https://docs.ganeti.org/docs/ganeti/3.0/html/man-ganeti-extstorage-interface.html). It _translates_ Ganeti operations to CSI operations, and stores returned metadata in `Metadata storage`. During requests, `ganeti-extstorage-csi` contacts `Metadata storage`.

## Metadata storage

Upon volume creation, CSI returns data, which is stored in `Metadata storage`. This data should be accessible on all nodes. For this, an etcd cluster is recommended to be set up across all nodes.

For testing/development purposes, a simple file based metadata storage is available, which stores metadata in files. This is just for development, not for production.

## TODO

* Make all operations as idempotent as possible.

## Install

Running simply `ganeti-extstorage-csi-install` will install the extstorage provider named `csi`. If you want a different name, pass it via environment variable `PROVIDER` when running the install script. This will actually populate folder `/usr/lib/ganeti-extstorage-csi/$PROVIDER` and configuration file `/etc/ganeti-extstorage-csi/$PROVIDER.env`. Edit the latter to setup CSI endpoint, metadata storage.

## Usage

When set up correctly, creating a ganeti instance with disk provided by this driver is simple, just run

```bash
# gnt-instance add -t ext --disk 0:size=10G,provider=<provider> ...
```

Set `<provider>` to the provider name (`csi` by default).

### TrueNAS-CSI settings

A TrueNAS csi driver can handle multiple NAS and each NAS may have multiple configurations, see https://github.com/dravanet/truenas-csi/tree/master/examples. ganeti-extstorage-csi has support for selecting the desired configuration. These are available as extstorage parameters:

- truenas_csi_nas
- truenas_csi_config

They are optional, and can be passed as options:

```bash
# gnt-instance add -t ext --disk 0:size=10G,provider=<provider>,truenas_csi_nas=xxx,truenas_csi_config=yyy ...
```
