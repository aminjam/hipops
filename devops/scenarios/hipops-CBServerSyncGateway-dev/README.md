## Running hipops-CBServerSyncGateway-dev
**(CBServer)couchbase-server + Sync-Gateway (CoreOS)**.
This scenario is build to [todo-lite demo](https://github.com/couchbaselabs/TodoLite-PhoneGap) and will:


1. create a couchbase server
2. deploy sync_gateway with nginx-proxy to host the sync endpoint at `cbsync-gateway-demo.com`

[Getting Started Guide & Setup](https://github.com/aminjam/hipops/wiki/Getting-Started#running-hipops-cbserversyncgateway-dev)

### Connecting with [TodoLite-PhoneGap](https://github.com/couchbaselabs/TodoLite-PhoneGap) with changing `js/index.js`:

  - `REMOTE_SYNC_URL` to be `cbsync-gateway-demo.com`

  - `FacebookInAppBrowser.settings.appId` with your FacebookAppId in with `
