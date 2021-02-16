# ganeti-extstorage-csi

Ganeti support for CSI through external storage interface.

WARNING: This software is still in development, some csi requirements are still missing, so it it not production ready.

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

To be fully CSI compliant:

- Invoke all controller operations if advertised (e.g. ControllerPublish*)
- Invoke all node operations if advertiset (e.g. NodeExpandVolume)
