# CLI Commands Reference

Dizz provides a comprehensive set of commands for managing your development workflow.

## Command Structure

All Dizz commands follow this pattern:

```bash
dizz [command] [subcommand] [flags] [args]
```

## Global Flags

These flags can be used with any command:

- `--help, -h` - Show help for the command
- `--version, -v` - Show version information
- `--config` - Path to configuration file (default: ~/.config/dizz/config.yaml)
- `--verbose` - Enable verbose output
- `--quiet, -q` - Suppress non-error output

## Core Commands

### `dizz`

Display help and overview of available commands.

```bash
dizz
```

### `dizz version`

Display the current version of Dizz.

```bash
dizz version
# Output: dizz version 1.0.0 (build abc123)
```

Flags:
- `--short` - Show only version number (e.g., "1.0.0")
- `--commit` - Show commit hash
- `--build-date` - Show build date

### `dizz help`

Show help for a specific command.

```bash
dizz help [command]
```

Examples:
```bash
dizz help build
dizz help install
```

## Project Management

### `dizz init`

Initialize a new Dizz project in the current directory.

```bash
dizz init [project-name]
```

Flags:
- `--template` - Use a specific template
- `--force` - Force initialization even if directory is not empty
- `--git` - Initialize git repository

Examples:
```bash
dizz init my-project
dizz init --template=web-app
dizz init --force
```

### `dizz install`

Install project dependencies.

```bash
dizz install [package...]
```

Flags:
- `--dev` - Install development dependencies only
- `--production` - Install production dependencies only
- `--save` - Save to package file
- `--global` - Install globally

Examples:
```bash
dizz install
dizz install --dev
dizz install react@latest --save
```

### `dizz build`

Build the project for production.

```bash
dizz build [target...]
```

Flags:
- `--env` - Build environment (default: production)
- `--minify` - Minify output
- `--output` - Output directory
- `--watch` - Watch for changes and rebuild

Examples:
```bash
dizz build
dizz build --env=development
dizz build --output=dist --minify
```

## Development

### `dizz dev`

Start development server with hot reload.

```bash
dizz dev
```

Flags:
- `--port` - Port number (default: 3000)
- `--host` - Host address (default: localhost)
- `--open` - Open browser automatically
- `--ssl` - Use HTTPS

Examples:
```bash
dizz dev
dizz dev --port=8080 --open
dizz dev --host=0.0.0.0
```

### `dizz test`

Run tests.

```bash
dizz test [pattern...]
```

Flags:
- `--watch` - Watch for changes and re-run tests
- `--coverage` - Generate coverage report
- `--verbose` - Show detailed test output
- `--timeout` - Test timeout in seconds

Examples:
```bash
dizz test
dizz test --coverage
dizz test --watch --verbose
```

### `dizz lint`

Run code linting.

```bash
dizz lint [files...]
```

Flags:
- `--fix` - Auto-fix issues where possible
- `--format` - Output format (json, table, etc.)
- `--config` - Lint configuration file

Examples:
```bash
dizz lint
dizz lint --fix
dizz lint --format=json
```

## Utility Commands

### `dizz clean`

Clean build artifacts and temporary files.

```bash
dizz clean
```

Flags:
- `--all` - Clean all caches
- `--dry-run` - Show what would be cleaned without actually cleaning

### `dizz doctor`

Check system and project health.

```bash
dizz doctor
```

This command checks:
- System dependencies
- Project configuration
- Common issues and fixes

### `dizz config`

Manage Dizz configuration.

```bash
dizz config [key] [value]
```

Subcommands:
- `get` - Get configuration value
- `set` - Set configuration value
- `list` - List all configuration
- `reset` - Reset to defaults

Examples:
```bash
dizz config get build.env
dizz config set build.env production
dizz config list
```

## Advanced Commands

### `dizz generate`

Generate code from templates.

```bash
dizz generate [template] [name]
```

Flags:
- `--template-dir` - Custom template directory
- `--output` - Output directory
- `--vars` - Template variables (key=value)

Examples:
```bash
dizz generate component MyComponent
dizz generate service AuthService --vars="path=auth,method=token"
```

### `dizz deploy`

Deploy the project.

```bash
dizz deploy [target]
```

Flags:
- `--env` - Deployment environment
- `--dry-run` - Show deployment plan without executing
- `--force` - Force deployment even if checks fail

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Invalid usage or arguments
- `3` - Network or I/O error
- `4` - Build failed
- `5` - Test failed

## Environment Variables

- `DIZZ_CONFIG` - Path to configuration file
- `DIZZ_ENV` - Environment (development, production, etc.)
- `DIZZ_LOG_LEVEL` - Log level (debug, info, warn, error)
- `DIZZ_NO_COLOR` - Disable colored output

For more detailed information about any command, use:

```bash
dizz help [command]
```