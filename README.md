# Quai-Sync
A tool for synchronizing Quai blockchain data.

## Description
Quai-Sync is a synchronization tool that helps you fetch and store Quai blockchain data efficiently. It uses a modified version of the Quai client library to ensure reliable data retrieval.

## Prerequisites
- Go 1.23.1 or higher
- Make build tools
- PostgreSQL database

## Usage
### Build from source:
```bash
git clone https://github.com/IPFS-Force/quai-syncer.git
cd quai-syncer
make build
```

### Run the tool:

```bash
Usage:
  quai-sync [flags]

Flags:
  -c, --config string   Config file path (default "config/config.toml")
  -h, --help           Show help information
  -v, --version        Show version information
```
Examples:
```bash
  # Run with default config
  $ quai-sync

  # Run with custom config file
  $ quai-sync --config=/path/to/config.toml

  # Show version
  $ quai-sync --version
```

The current file uses the github.com/GalaxiesCN/go-quai v0.39.4 version, because the github.com/dominant-strategies/go-quai v0.39.4 version has problems with the ethclient method.