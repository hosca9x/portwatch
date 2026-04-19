# portwatch

Lightweight CLI daemon that monitors open ports and alerts on unexpected changes.

## Installation

```bash
go install github.com/yourusername/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/portwatch.git && cd portwatch && go build -o portwatch .
```

## Usage

Start the daemon with default settings (scans every 60 seconds):

```bash
portwatch start
```

Specify a custom scan interval and alert on any change:

```bash
portwatch start --interval 30s --notify
```

Define a baseline of expected open ports:

```bash
portwatch baseline --ports 22,80,443
```

View current port status:

```bash
portwatch status
```

When an unexpected port opens or closes, `portwatch` logs the event and optionally sends a desktop or webhook notification.

### Example Output

```
[2024-03-12 14:02:11] INFO  Baseline loaded: 22, 80, 443
[2024-03-12 14:02:41] ALERT New port detected: 8080 (process: python3, pid: 4821)
[2024-03-12 14:03:11] INFO  Scan complete. No changes detected.
```

## Configuration

`portwatch` looks for a config file at `~/.config/portwatch/config.yaml`:

```yaml
interval: 30s
baseline:
  - 22
  - 80
  - 443
notify:
  webhook: "https://hooks.example.com/alert"
```

## License

MIT © 2024 Your Name