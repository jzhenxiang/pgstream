# pgstream

Lightweight CDC tool that streams Postgres WAL changes to Kafka or webhook endpoints with minimal config.

## Features

- Stream Postgres WAL changes in real-time via logical replication
- Multiple output targets: Kafka, webhooks, or stdout
- Minimal configuration required
- Low memory footprint and CPU usage
- Automatic reconnection and error handling

## Installation

```bash
go install github.com/yourusername/pgstream@latest
```

Or download pre-built binaries from the [releases page](https://github.com/yourusername/pgstream/releases).

## Usage

Create a `config.yaml`:

```yaml
postgres:
  host: localhost
  port: 5432
  database: mydb
  user: postgres
  password: secret
  
output:
  type: kafka
  brokers:
    - localhost:9092
  topic: postgres-changes
```

Run pgstream:

```bash
pgstream --config config.yaml
```

For webhook output:

```yaml
output:
  type: webhook
  url: https://your-endpoint.com/changes
  headers:
    Authorization: "Bearer your-token"
```

## Prerequisites

- PostgreSQL 10+ with logical replication enabled (`wal_level = logical`)
- Replication slot and publication configured on source database

## License

MIT