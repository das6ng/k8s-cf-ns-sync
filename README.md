# k8s-cf-ns-sync
Simply sync NS record with k8s ingress resource annotion.

# configuration

env vars:

- `MONITOR_NS`: monitoring k8s namespaces

- `CLOUDFLARE_API_TOKEN`: cloudflare `api_token`

- `CLOUDFLARE_ZONE_NAME`: cloudflare managed DNS name

# ingress annotation

- `"cf-ns-sync/name"`: DNS A record name

- `"cf-ns-sync/value"`: DNS A record content
