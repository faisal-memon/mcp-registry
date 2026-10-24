# MCP Registry

`mcp-registry` builds an in-memory catalogue of MCP servers from running Docker containers. It is intended to replace a hand-maintained registry document, but its first release is designed to run beside the existing static registry for comparison.

It does not store state. On startup and at every reconciliation interval, it lists running labelled containers, asks each MCP endpoint for its tools, and atomically publishes the successful results. If Docker is briefly unavailable, the last successful snapshot remains available. If an MCP endpoint cannot be reached, it remains in the snapshot as `enabled: false` and will be retried later.

## Container contract

Label an MCP container with all of these values:

```yaml
labels:
  mcp.enabled: "true"
  mcp.port: "8787"
  mcp.path: /mcp # optional; this is the default
```

The registry uses the container name as the catalogue name and the network host. That name must be resolvable from the registry container through the shared MCP Docker network.

The defaults are neutral and can be replaced without rebuilding the service:

| Setting | Default |
| --- | --- |
| `MCP_LABEL_ENABLED_KEY` | `mcp.enabled` |
| `MCP_LABEL_ENABLED_VALUE` | `true` |
| `MCP_LABEL_PORT_KEY` | `mcp.port` |
| `MCP_LABEL_PATH_KEY` | `mcp.path` |
| `MCP_DEFAULT_PATH` | `/mcp` |

## Running it

The registry needs read-only access to Docker's container-list API. In deployment, route this through a restricted Docker socket proxy rather than mounting the Docker socket directly.

```sh
docker run --rm \
  --network mcp_net \
  -e DOCKER_HOST=tcp://docker-socket-proxy:2375 \
  -e RECONCILE_INTERVAL=30s \
  -p 8080:8080 \
  ghcr.io/faisal-memon/mcp-registry:latest
```

For a first manual comparison, request:

```sh
curl http://mcp-registry-auto:8080/registry.json
curl http://mcp-registry-auto:8080/healthz
```

`GET /registry.json` returns the existing `servers` registry shape plus each server's discovered `tools`. This service is not wired into the RAG agent in this first cut.

## Development

```sh
make test
make lint
```

The initial implementation reconciles on a short interval. Docker event handling and deployment wiring are intentionally follow-up work, after this endpoint has been compared with the static registry.
