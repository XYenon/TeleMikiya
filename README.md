# TeleMikiya

TeleMikiya is a hybrid message search tool for Telegram that combines semantic similarity search with full-text search capabilities, providing more accurate and comprehensive search results.

## Features

- 🔄 Automatic Telegram message syncing and indexing
- 🔍 Hybrid search combining:
  - Semantic similarity search using vector embeddings
  - Full-text search powered by PGroonga
- 🤖 Multiple embedding providers support:
  - [Ollama](https://ollama.ai/)
  - [OpenAI](https://platform.openai.com/docs/guides/embeddings)
- 💬 Both CLI and Telegram Bot interfaces

## Requirements

- PostgreSQL 15+ with:
  - [VectorChord](https://docs.vectorchord.ai/) extension for vector similarity search
  - [PGroonga](https://pgroonga.github.io/) extension for full-text search
- Either [Ollama](https://ollama.ai/) or OpenAI API access

## Configuration

1. Create a configuration file:

```bash
# Copy example config to one of:
# - ./config.toml
# - $XDG_CONFIG_HOME/telemikiya/config.toml
# - /etc/telemikiya/config.toml
cp config.example.toml config.toml
```

2. Edit the configuration file with required settings:

- Telegram API credentials (from https://my.telegram.org)
- Database connection details
- Text embedding service configuration

## Database Setup

1. Install required PostgreSQL extensions:

- Install VectorChord following the [official installation guide](https://docs.vectorchord.ai/vectorchord/getting-started/installation.html)
- Install PGroonga following the [official installation guide](https://pgroonga.github.io/install/)

2. Create database and user:

```sql
-- Connect as postgres user
sudo -u postgres psql

-- Create user
CREATE USER telemikiya WITH PASSWORD 'your_password';

-- Create database
CREATE DATABASE telemikiya OWNER telemikiya;

-- Connect to the new database
\c telemikiya

-- Create extensions
CREATE EXTENSION IF NOT EXISTS vchord CASCADE;
CREATE EXTENSION IF NOT EXISTS pgroonga;

-- Create schemas
CREATE SCHEMA user_session AUTHORIZATION telemikiya;
CREATE SCHEMA bot_session AUTHORIZATION telemikiya;
```

3. Migrate database schema:

```bash
# First-time setup
telemikiya db migrate

# When embedding dimensions change, use --allow-clear-embedding flag
telemikiya db migrate --allow-clear-embedding
```

## Usage

### Start Services

```bash
# Start all services (observer, embedding, and bot)
telemikiya run

# Start specific services only
telemikiya run --observer=false --embedding=true --bot=false
```

### Search Messages

```bash
# Basic search
telemikiya search "how to use Docker"

# Specify result count
telemikiya search --count 20 "recommend a movie"

# Search in specific dialog
telemikiya search --dialog-id 123456789 "meeting notes"

# Search by time range
telemikiya search --start-time "2024-01-01 00:00:00" "happy new year"
```

### Use Telegram Bot

Send `/search` command to your bot:

```
/search how to use Docker
```

### Debug Mode

Enable debug logging with `-D` or `--debug`:

```bash
telemikiya -D run
```

### Alternative Config File

Use `-C` or `--config` to specify config file path:

```bash
telemikiya -C /path/to/config.toml run
```
