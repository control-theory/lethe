# Lethe

*Named for the river of forgetfulness in Greek mythology.*

Local graph-based memory for Claude. Memories reinforce when used and fade when ignored.

## What it does

- **Remember** facts, people, projects, preferences across conversations
- **Recall** by keyword — FTS5 full-text search with relevance ranking
- **Decay** — unused memories fade over time (ACT-R exponential decay, configurable half-life)
- **Pin** — important memories never decay
- **Graph** — memories connect to each other via typed relations
- **Dashboard** — web UI to browse, search, pin, and delete memories

## Install

### Claude Desktop (MCPB bundle)

Download `lethe-0.1.0.mcpb` from Releases. Double-click to install in Claude Desktop.

### Build the MCPB bundle from source

```bash
# Install the MCPB packaging tool
npm install -g @anthropic-ai/mcpb

# Cross-compile + stage the bundle
make bundle

# Pack into a .mcpb file
mcpb pack bundle/
```

This produces `lethe-0.1.0.mcpb` — a single file you can double-click to install in Claude Desktop. It contains binaries for macOS (arm64 + amd64), Linux (amd64), and Windows (amd64).

### Claude Code

```bash
# Build from source
make build

# Add to Claude Code
claude mcp add lethe -- ./bin/lethe --stdio
```

### Manual (Claude Desktop JSON config)

```json
{
  "mcpServers": {
    "lethe": {
      "command": "/path/to/lethe",
      "args": ["--stdio"]
    }
  }
}
```

## Usage

Once installed, Claude will automatically have access to these tools:

| Tool | What it does |
|---|---|
| `remember` | Store a fact, preference, or entity |
| `recall` | Search memories by keyword |
| `forget` | Soft-forget or permanently delete |
| `pin` / `unpin` | Toggle decay immunity |
| `link` / `unlink` | Create/remove relations between memories |
| `get_context` | Fetch the subgraph around a memory |
| `update_memory` | Edit an existing memory |
| `list_recent` | Recently created/reinforced memories |
| `stats` | Memory system statistics |

Claude also gets three **resources** that auto-surface at session start:
- `lethe://context/active` — top 20 active memories
- `lethe://stats/summary` — compact stats
- `lethe://schema/types` — available node kinds and edge types

## Dashboard

The leader process (first `lethe` instance) automatically starts a localhost web dashboard. Open it with:

```bash
lethe dashboard
```

Features: graph visualization, list view, timeline, fading memories section, pin/delete/edit, export/import.

## Data

- **Database:** `~/Library/Application Support/lethe/memory.db` (macOS), `~/.config/lethe/memory.db` (Linux)
- **Format:** SQLite with WAL mode. Multi-process safe.
- **Export:** `GET /api/export` from dashboard, or use the dashboard UI
- **Import:** `POST /api/import` with JSON body

## Decay model

Memories decay exponentially: `S(t) = S₀ × e^(-λ × Δt)` where `λ = ln(2) / half_life_days`.

- Default half-life: **30 days**
- Reinforcement on retrieval: **+1.0** (capped at 10.0)
- Pinned memories never decay
- Strength < 0.1 = "fading" (excluded from default search)
- Nothing is ever auto-deleted

## Build

```bash
make build          # local binary → bin/lethe
make test           # run all tests
make lint           # go vet
make build-all      # cross-compile for darwin/linux/windows
make bundle         # build-all + copy to bundle/server/
```

## Tech stack

- Go 1.23+
- SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- MCP SDK: `github.com/modelcontextprotocol/go-sdk`
- Single static binary per platform
