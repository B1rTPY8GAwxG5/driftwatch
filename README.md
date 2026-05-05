# driftwatch

Detect configuration drift between deployed services and their declared infrastructure-as-code definitions.

---

## Installation

```bash
go install github.com/yourusername/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/driftwatch.git && cd driftwatch && go build ./...
```

---

## Usage

Point `driftwatch` at your IaC definitions and a target environment to scan for drift:

```bash
driftwatch scan --config ./infra/config.yaml --env production
```

Example output:

```
[DRIFT] service: api-gateway
  expected: instance_type = t3.medium
  actual:   instance_type = t3.large

[OK] service: auth-service
[OK] service: worker

2 services checked. 1 drift(s) detected.
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to IaC definition file | `./config.yaml` |
| `--env` | Target environment to inspect | `production` |
| `--output` | Output format: `text`, `json` | `text` |
| `--fail-on-drift` | Exit with non-zero code if drift is found | `false` |

---

## Configuration

```yaml
# config.yaml
services:
  - name: api-gateway
    instance_type: t3.medium
    replicas: 3
  - name: auth-service
    instance_type: t3.small
    replicas: 2
```

---

## License

MIT © 2024 driftwatch contributors