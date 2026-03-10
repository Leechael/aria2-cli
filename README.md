# aria2-cli

Command-line client for [aria2](https://aria2.github.io/) JSON-RPC interface.

## Install

```bash
go install github.com/leechael/aria2-cli-skills/cmd/aria2-cli@latest
```

Or build from source:

```bash
make build
```

## Configuration

Via environment variables:

| Variable | Default | Description |
|---|---|---|
| `ARIA2_RPC_HOST` | `localhost` | aria2 RPC host |
| `ARIA2_RPC_PORT` | `6800` | aria2 RPC port |
| `ARIA2_RPC_SECRET` | _(empty)_ | RPC secret token |
| `ARIA2_RPC_SECURE` | `false` | Use HTTPS (`true` or `1`) |

## Usage

```bash
aria2-cli [--json|--plain] <command> [args...]
```

`--json` outputs machine-readable JSON. `--plain` (default) outputs human-readable text. Hints and progress go to stderr; data goes to stdout.

### Download management

```bash
# Add download
aria2-cli add https://example.com/file.zip

# Add with options
aria2-cli add https://example.com/file.zip dir=/tmp max-connection-per-server=4

# Add multiple URIs (same file, mirrors)
aria2-cli add https://mirror1.com/file.zip https://mirror2.com/file.zip

# Pause / resume
aria2-cli pause <gid>
aria2-cli resume <gid>
aria2-cli pause-all
aria2-cli resume-all

# Remove
aria2-cli remove <gid>
aria2-cli force-remove <gid>
```

### Status

```bash
# Single download status
aria2-cli status <gid>

# List downloads
aria2-cli list              # all
aria2-cli list active       # active only
aria2-cli list waiting      # waiting only
aria2-cli list stopped      # stopped only

# Global statistics
aria2-cli stat

# aria2 version
aria2-cli version
```

### Download details

```bash
aria2-cli get-files <gid>
aria2-cli get-uris <gid>
aria2-cli get-peers <gid>
aria2-cli get-servers <gid>
```

### Options

```bash
# Per-download options
aria2-cli get-option <gid>
aria2-cli set-option <gid> max-download-limit=100K

# Global options
aria2-cli get-global-option
aria2-cli set-global-option max-overall-download-limit=1M
```

### Session

```bash
aria2-cli save-session
aria2-cli purge
aria2-cli remove-result <gid>
aria2-cli shutdown
aria2-cli force-shutdown
```

### JSON output for scripting

```bash
# Get active GIDs as JSON
aria2-cli --json list active | jq '.[].gid'

# Check global download speed
aria2-cli --json stat | jq '.downloadSpeed'
```

## Development

```bash
make test     # Run tests
make lint     # vet + fmt check
make cross    # Cross-compile (linux/darwin, amd64/arm64)
make clean    # Remove build artifacts
```

## License

MIT
