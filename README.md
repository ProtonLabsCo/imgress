* create a intermediary page for "uploaded"
* save info into postgres to delete old images
* keep total number of MBs saved in history (postgres)
* keep total number of unique users (postgres)
* dockerize the app and use k8s
* use rabbitmq to send images to be compressed to compressor service
* write unit tests for all methods
* grafana + promotheus
* serve over https, with lets encrypt
* general refactoring

* domain name: middle-out compression from silicon valley :)))

### Run this command
```
    CGO_CFLAGS_ALLOW="-Wl,(-framework|CoreFoundation)"
    export CGO_CFLAGS_ALLOW=".*"
```
