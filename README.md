# steam

A service for indexing and retrieving Steam game metadata. It maintains an optimized in-memory database of games with automated daily updates and real-time discovery for unindexed items.

## Features

- **Quick Search**: Token-based search for games with results sorted by relevance.
- **Deep Lookup**: Retrieves game metadata including DLC lists and parent game information.
- **On-Demand Discovery**: Automatically fetches and caches data for games not yet present in the local index.
- **Automated Updates**: Periodically syncs with the Steam registry to ensure data accuracy.
- **Optimized Performance**: Utilizes efficient serialization and memory management for minimal latency.

## Configuration

The service is configured via environment variables:

| Variable | Description | Status |
|----------|-------------|--------|
| `STEAM_KEY` | Steam Web API Key | Required |
| `PORT` | API listening port (default: 8080) | Optional |

## API Reference

### GET /lookup
Fetches metadata for a specific game.

**Parameters:**
- `id`: The Steam Game ID.

**Response Formats:**
- **Games**: `[id, "Name", [[dlc_id, "DLC Name"], ...]]`
- **DLCs**: `[id, "Name", [parent_id, "Parent Name"]]`

### GET /search
Searches the game registry.

**Parameters:**
- `q`: Search query (minimum 2 characters).

**Response Format:**
- `[[id, "Name"], [id, "Name"], ...]`
