# SSH Config Support

This package now supports reading SSH configuration from `~/.ssh/config` for SFTP connections, making it easier to connect to remote servers.

## Features

- **Read `~/.ssh/config`**: Automatically loads SSH configuration from the standard location
- **Multiple URL formats**: Supports various SSH/SFTP URL formats
- **Wildcard matching**: Supports SSH config host patterns like `*.example.com`
- **Credential management**: Automatically loads SSH keys and credentials from config

## Supported URL Formats

### 1. SFTP with full URL
```go
// Standard SFTP URL
url := "sftp://user@example.com:22/path/to/file"
```

### 2. SSH shorthand format
```go
// ssh: prefix (uses SSH config for the hostname)
url := "ssh:myserver"

// sftp: prefix
url := "sftp:myserver"
```

### 3. User@Host format (SSH-style)
```go
// Basic format
url := "user@example.com"

// With path
url := "user@example.com:/path/to/file"

// With custom port and path
url := "user@example.com:2222/path/to/file"

// With absolute path
url := "user@example.com:/absolute/path"
```

## SSH Config Example

Your `~/.ssh/config` file might look like this:

```ssh-config
# Production server
Host prod
    HostName 192.168.1.100
    User admin
    Port 2222
    IdentityFile ~/.ssh/id_rsa_prod

# Development servers
Host *.dev.example.com
    User developer
    Port 22
    IdentityFile ~/.ssh/id_rsa_dev
    ServerAliveInterval 60

# Default settings for all hosts
Host *
    ServerAliveInterval 30
    ServerAliveCountMax 3
```

## Usage Examples

### Basic SFTP Connection

```go
import (
    "github.com/qiangyt/go-comm/v2"
    "github.com/spf13/afero"
)

// Using SSH config hostname
file, err := comm.NewFile(afero.NewOsFs(), "ssh:prod/data/file.txt", nil, 30*time.Second)
if err != nil {
    log.Fatal(err)
}

content, err := file.Download()
if err != nil {
    log.Fatal(err)
}
```

### Using SSH-style Format

```go
// The package will automatically read ~/.ssh/config to get:
// - hostname (192.168.1.100 from HostName)
// - username (admin from User)
// - port (2222 from Port)
// - SSH key (from IdentityFile)

file, err := comm.NewFile(
    afero.NewOsFs(),
    "admin@prod:/path/to/file",
    nil, // credentials loaded from SSH config
    30*time.Second,
)
```

### Programmatic SSH Config Usage

```go
// Load SSH config explicitly
sshConfig, err := comm.LoadSSHConfig()
if err != nil {
    log.Fatal(err)
}

// Get config for a specific host
hostConfig := sshConfig.GetHostConfig("prod")
if hostConfig != nil {
    fmt.Printf("Hostname: %s\n", hostConfig.HostName)
    fmt.Printf("User: %s\n", hostConfig.User)
    fmt.Printf("Port: %d\n", hostConfig.Port)
}

// Convert to credentials
creds := comm.SSHConfigToCredentials(hostConfig)
```

### Parse SSH URL

```go
// Parse various SSH URL formats
protocol, user, hostname, path, port, err := comm.ParseSSHURL("user@example.com:2222/data")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Protocol: %s\n", protocol) // "ssh"
fmt.Printf("User: %s\n", user)         // "user"
fmt.Printf("Hostname: %s\n", hostname) // "example.com"
fmt.Printf("Path: %s\n", path)         // "/data"
fmt.Printf("Port: %d\n", port)         // 2222
```

### Get Credentials from SSH Config

```go
// This function:
// 1. Parses the SSH URL
// 2. Loads ~/.ssh/config
// 3. Finds matching host configuration
// 4. Returns credentials with SSH key loaded

creds, hostname, port, path, err := comm.NewCredentialsFromSSH("admin@prod:/data/file.txt")
if err != nil {
    log.Fatal(err)
}

// Use the credentials
url := comm.BuildSSHURL("sftp", creds.User, hostname, port, path)
file, err := comm.NewRemoteFile(url, creds, 30*time.Second)
```

## Supported SSH Config Directives

The following SSH config directives are currently supported:

- `Host` - Host pattern matching
- `HostName` - Real hostname or IP address
- `User` - Username for authentication
- `Port` - Port number (default: 22)
- `IdentityFile` - Path to private key file
- `PreferredAuthentications` - Authentication methods
- `ProxyJump` - Jump host configuration
- `ProxyCommand` - Proxy command
- `ForwardAgent` - SSH agent forwarding
- `Compression` - Enable/disable compression
- `ServerAliveInterval` - Keep-alive interval
- `ServerAliveCountMax` - Keep-alive retry count
- `StrictHostKeyChecking` - Host key verification
- `ConnectTimeout` - Connection timeout
- And more...

## Pattern Matching

The SSH config parser supports wildcard patterns:

- `*` - Matches zero or more characters
- `?` - Matches exactly one character

Examples:
- `*.example.com` - Matches any subdomain of example.com
- `server*` - Matches server1, server2, server-prod, etc.
- `Host *` - Matches all hosts (used for default settings)

## Notes

- If `~/.ssh/config` doesn't exist, the parser returns an empty configuration without error
- SSH config is loaded automatically when using SSH-style URLs
- Private keys are read from the path specified in `IdentityFile`
- Values explicitly provided in the URL override SSH config settings
- This feature is primarily designed for SFTP connections
