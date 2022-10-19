* create a intermediary page for "uploaded"
* serve over https, with lets encrypt
* dockerize the app and use k8s
* use rabbitmq to send images to be compressed to compressor service
* run both webapp and cron in k8s
* save compressed files into aws s3 bucket (or rather wasabi)
* hide link in href from the user
* write unit tests for all methods
* grafana + promotheus
* general refactoring
* keep total number of MBs saved in history
* keep total number of unique users

* domain name: middle-out compression from silicon valley :)))

### Run this command
```
    CGO_CFLAGS_ALLOW="-Wl,(-framework|CoreFoundation)"
    export CGO_CFLAGS_ALLOW=".*"
```
