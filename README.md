# MCP Registry

`mcp-registry` automatically builds an in-memory registry of MCP servers from labelled Docker containers. It replaces a hand-maintained server list with the tools each running MCP service currently advertises.

On startup, and again at a configurable interval, it:

1. Lists running Docker containers.
2. Selects containers carrying the configured MCP labels.
3. Connects to each container’s MCP endpoint and requests its tool list.
4. Serves the resulting registry at `GET /registry.json`.

The registry keeps no persistent state. If Docker is temporarily unavailable, it continues serving its last successful snapshot. If one MCP endpoint cannot be reached, that service remains listed as `enabled: false` and is retried during the next reconciliation.

## Run the registry

The registry must share a Docker network with the MCP services it discovers. It also needs read-only access to Docker’s container-list API. In deployment, use a restricted Docker socket proxy instead of mounting Docker’s full socket directly.

```yaml
services:
  mcp-registry:
    image: ghcr.io/faisal-memon/mcp-registry:latest
    environment:
      DOCKER_HOST: tcp://docker-socket-proxy:2375
      RECONCILE_INTERVAL: 30s
    networks:
      - mcp_net
    ports:
      - "8080:8080" # Optional: expose only when the host needs access.

networks:
  mcp_net:
    external: true
```

Check that it is running and inspect its snapshot:

```sh
curl http://mcp-registry:8080/healthz
curl http://mcp-registry:8080/registry.json
```

The registry returns the existing `servers` document shape, including the tools discovered from each service. This first release runs alongside the static registry; it does not yet change the RAG agent’s configuration.

## Make an MCP service discoverable

Add these labels to every MCP service’s Compose definition:

```yaml
services:
  muni-mcp:
    image: example/muni-mcp:latest
    networks:
      - mcp_net
    labels:
      mcp.enabled: "true"
      mcp.port: "8787"
      mcp.path: "/mcp" # Optional: /mcp is the default.
```

The registry uses the container name, here `muni-mcp`, as both the server name and network host. That name must resolve from the registry container on the shared network. `mcp.port` is required; `mcp.path` is optional.

### Change the label contract

The defaults are generic, and may be changed through the registry container’s environment without rebuilding it:

| Setting | Default | Purpose |
| --- | --- | --- |
| `MCP_LABEL_ENABLED_KEY` | `mcp.enabled` | Label that opts a container into discovery. |
| `MCP_LABEL_ENABLED_VALUE` | `true` | Required value of the opt-in label. |
| `MCP_LABEL_PORT_KEY` | `mcp.port` | Label containing the MCP endpoint port. |
| `MCP_LABEL_PATH_KEY` | `mcp.path` | Label containing the MCP endpoint path. |
| `MCP_DEFAULT_PATH` | `/mcp` | Path used when the container omits `mcp.path`. |
| `LISTEN_ADDR` | `:8080` | Registry HTTP listen address. |
| `RECONCILE_INTERVAL` | `30s` | How often the registry rescans containers and endpoints. |

## Development

```sh
make test
make lint
```

Docker event handling and deployment wiring are intentional follow-up work, after this endpoint has been compared with the static registry.
