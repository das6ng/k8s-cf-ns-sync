# Intro

Simply sync DNS record from k8s ingress resource annotion to `cloudflare.com`.

# Usage

Deploy the image to your cluster, and give it authorization to "get/list/watch" your ingress 
resources in the target namespaces.

image: `ghcr.io/das6ng/k8s-cf-ns-sync:latest`

deploy example: [link](https://github.com/das6ng/k8s-cf-ns-sync/blob/main/deployment-example.yaml)

## Cli Flags

```txt
GLOBAL OPTIONS:
   --mode string                          [exclude]|include (default: "exclude")
   --exclude string [ --exclude string ]  monitor all namespace except specified by this flag (default: "kube-system", "kube-public", "kube-node-lease")
   --include string [ --include string ]  monitor namespace specified by this flag (default: "default")
   --cloudflare-zone string               DNS zone name managed by Cloudflare.com
   --cloudflare-api-token string          api-token of Cloudflare.com, need ZONE-EDIT ZONE-READ access to specified Zone
   --help, -h                             show help
   --version, -v                          print the version
```

- `--mode <mode>`: specify monitor `namespace` mode, can be either `exclude`(default) or `include`.

- `--exclude <namespace> [--exclude <namespace>]`: in `exclude` mode, monitor all namespace except specified by this flag (default: `kube-system`, `kube-public`, `kube-node-lease`).

- `--include <namespace> [--include <namespace>]`: in `include` mode, monitor namespace specified by this flag (default: `default`).

- `--cloudflare-zone <zone_name>`: DNS zone name managed by [Cloudflare](https://cloudflare.com/)

- `--cloudflare-api-token <api_token>`: cloudflare `api_token`

    The [api token](https://dash.cloudflare.com/profile/api-tokens) *MUST* have the following `Permissions` on your target zone:

    ```
    Zone    DNS    Read
    Zone    DNS    Edit
    ```

## ENV vars


- `LOG_LEVEL`: running log level, should be `DEBUG/INFO/WARN/ERROR`


## Ingress annotation

- `"cf-ns-sync/name"`: DNS A record name

- `"cf-ns-sync/value"`: DNS A record content

Example:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-ingress
  annotations:
    "cf-ns-sync/name": test01.abc.com
    "cf-ns-sync/value": 191.168.1.99
spec:
  rules:
  # ...
```
