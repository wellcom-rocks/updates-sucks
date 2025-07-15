# Updates-Sucks CLI

A command-line tool for automating software version monitoring for DevOps engineers and system administrators.

## Features

- **Automated Version Scanning**: Monitors multiple repositories for new versions
- **Multiple Version Schemes**: Supports SemVer, CalVer, and string-based versioning
- **CI/CD Integration**: Standardized exit codes for pipeline automation
- **Authentication Support**: Handles private repositories with token and SSH authentication
- **Flexible Output**: Human-readable and JSON output formats
- **Configurable**: JSON-based configuration with extensive customization options

## Installation

### Direct Installation
```bash
go install github.com/wellcom-rocks/updates-sucks@latest
```

### Build from Source

```bash
go build -o updates-sucks
```

### Usage

```bash
# Scan all repositories
./updates-sucks scan

# Scan specific repository
./updates-sucks scan "Kubernetes"

# Use custom configuration file
./updates-sucks scan --file /path/to/config.json

# JSON output for automation
./updates-sucks scan --format json

# Verbose output for debugging
./updates-sucks scan --verbose

# Quiet output for CI/CD
./updates-sucks scan --quiet
```

## Configuration

Create a `repos.json` file in your project directory:

```json
{
  "repositories": [
    {
      "name": "Kubernetes",
      "type": "git",
      "url": "https://github.com/kubernetes/kubernetes.git",
      "currentVersion": "v1.28.0",
      "versioning": {
        "scheme": "semver",
        "ignorePrefix": "v"
      }
    },
    {
      "name": "Private Repository",
      "type": "git",
      "url": "git@github.com:company/private-repo.git",
      "currentVersion": "1.5.0",
      "auth": {
        "type": "token",
        "envVariable": "GITHUB_TOKEN"
      }
    }
  ]
}
```

### Configuration Options

- **`name`**: Human-readable name for the repository
- **`type`**: Repository type (currently only `"git"` is supported)
- **`url`**: Repository URL (HTTPS or SSH)
- **`currentVersion`**: Current version in use
- **`versioning`** (optional):
  - **`scheme`**: Version scheme (`"semver"`, `"calver"`, `"string"`)
  - **`ignorePrefix`**: Prefix to ignore when comparing versions (e.g., `"v"`)
- **`auth`** (optional):
  - **`type`**: Authentication type (`"token"` or `"ssh"`)
  - **`envVariable`**: Environment variable containing the token/key path

## Exit Codes

- **`0`**: Success, no updates available
- **`1`**: Success, updates found
- **`2`**: Configuration error
- **`3`**: Scan error (network, authentication, etc.)

## Examples

### Basic Usage

```bash
# Check for updates
./updates-sucks scan

# Output:
# Scanning 3 repositories...
#
# - Kubernetes: UP-TO-DATE (Current: v1.28.0)
# - Docker: NEW VERSION FOUND! (Current: v24.0.0 -> Latest: v24.0.5)
# - Prometheus: UP-TO-DATE (Current: v2.45.0)
#
# Scan finished. Updates available for 1 repository(ies).
```

### JSON Output

```bash
./updates-sucks scan --format json
```

```json
{
  "summary": {
    "total": 3,
    "upToDate": 2,
    "updatesAvailable": 1,
    "errors": 0
  },
  "repositories": [
    {
      "name": "Kubernetes",
      "status": "UP_TO_DATE",
      "currentVersion": "v1.28.0",
      "latestVersion": "v1.28.0"
    },
    {
      "name": "Docker",
      "status": "UPDATE_AVAILABLE",
      "currentVersion": "v24.0.0",
      "latestVersion": "v24.0.5"
    }
  ]
}
```

### CI/CD Integration

```bash
#!/bin/bash
./updates-sucks scan --quiet --format json

case $? in
    0)
        echo "All repositories are up to date"
        ;;
    1)
        echo "Updates are available!"
        # Trigger notification or further processing
        ;;
    2)
        echo "Configuration error"
        exit 1
        ;;
    3)
        echo "Scan failed"
        exit 1
        ;;
esac
```

## Authentication

### GitHub Token

```bash
export GITHUB_TOKEN="your-token-here"
./updates-sucks scan
```

### SSH Key

```bash
export SSH_KEY_PATH="/path/to/your/ssh/key"
./updates-sucks scan
```

## Supported Version Schemes

### Semantic Versioning (SemVer)
- Format: `MAJOR.MINOR.PATCH` (e.g., `1.2.3`, `v2.0.0`)
- Supports pre-release and build metadata
- Default scheme if not specified

### Calendar Versioning (CalVer)
- Format: `YYYY.MM.MICRO` (e.g., `2024.05.1`)
- Useful for date-based releases

### String Versioning
- Lexicographic comparison
- Fallback for non-standard versioning schemes

## Performance

- Uses `git ls-remote` for efficient tag fetching without cloning
- Lightweight and fast for CI/CD pipeline integration
- Minimal memory footprint

## Contributing

This project follows the specification detailed in `spec.md`. Please refer to the specification for architectural decisions and implementation details.

## License

This project is licensed under the MIT License.
