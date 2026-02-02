# Dizz CLI Documentation

Progress-aware CLI tool that tells you what to work on next.

## Installation

```bash
# Linux/macOS
curl -sSL https://dizz.shitworks.co/install.sh | bash

# Windows
powershell -c "irm https://dizz.shitworks.co/install.ps1 | iex"
```

## Quick Start

```bash
# Initialize in your project
dizz init

# See what needs attention
dizz log

# Quick status check
dizz status

# Save project state
dizz snapshot

# List all snapshots
dizz list

# Resume work after time away
dizz resume
```

## Commands

### `dizz init`
Initialize dizz in current directory.

### `dizz log`
Analyze codebase and show what needs attention:
- **Planned**: Functions with TODOs
- **Unstable**: Code changing too frequently  
- **Unused**: Functions declared but never called
- **Abandoned**: Old, unused code

Flags:
- `--all, -a`: Show all symbols including active ones
- `--verbose, -v`: Show detailed analysis

### `dizz status`
Quick project health check with visual indicators.

### `dizz snapshot`
Create immutable snapshot of current project state.
Flags:
- `--auto`: Automatic snapshot (called by git hook)

### `dizz list`
Show all saved snapshots with timestamps and metadata.
Flags:
- `--format`: Output format (json|table)

### `dizz resume`
Quick context after being away from the project.
Flags:
- `--days N`: Show context for last N days

### `dizz version`
Show current version and build information.
Flags:
- `--verbose`: Show detailed build information

### `dizz upgrade`
Upgrade to the latest version automatically.
Flags:
- `--force`: Force upgrade even if already latest

### `dizz intent`
Manage intentional TODO markers and track planned work.
Subcommands:
- `add [@symbol] [description]`: Add intent marker
- `list`: Show all intent markers
- `remove [@symbol]`: Remove intent marker
- `complete [@symbol]`: Mark intent as completed

## How It Works

1. **Static Analysis**: Parses code to find functions and calls
2. **TODO Detection**: Scans for TODO/FIXME comments  
3. **State Scoring**: Combines signals to determine function state
4. **Smart Output**: Shows exactly what needs attention

- No network calls
- No tokens consumed
- Runs in milliseconds

## Configuration

`.dizz/config.json`:
```json
{
  "project_name": "my-project",
  "root_path": ".",
  "exclude": ["vendor/**", "node_modules/**", ".git/**", ".dizz/**"]
}
```

## Symbol States

- **`active`**: Used and stable
- **`planned`**: Has TODO markers, needs implementation
- **`unstable`**: High churn, changing too frequently
- **`unused`**: Declared but never called
- **`abandoned`**: Old, unused, and stale