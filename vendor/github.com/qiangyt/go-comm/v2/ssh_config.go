package comm

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/goodsru/go-universal-network-adapter/models"
	"github.com/pkg/errors"
)

// parseIntDefault parses string to int with default value
func parseIntDefault(s string, defaultValue int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return defaultValue
}

// SSHHostConfig represents SSH configuration for a specific host
type SSHHostConfig struct {
	Host                 string
	HostName             string
	User                 string
	Port                 int
	IdentityFile         string
	IdentityFiles        []string
	PreferredAuth        []string
	ProxyJump            string
	ProxyCommand         string
	ForwardAgent         bool
	Compression          bool
	ServerAliveInterval  int
	ServerAliveCountMax  int
	StrictHostKeyCheck   string
	UserKnownHostsFile   string
	ConnectTimeout       int
	PasswordAuth         bool
	PubkeyAuth           bool
	KbdInteractiveAuth   bool
	RekeyLimit           string
	SendEnv              []string
	SetEnv               map[string]string
	RequestTTY           string
	RemoteCommand        string
	LocalForward         []string
	RemoteForward        []string
	DynamicForward       []string
	ExitOnForwardFailure bool
	ControlMaster        string
	ControlPath          string
	ControlPersist       string
}

// SSHConfig represents the parsed SSH configuration
type SSHConfig struct {
	hosts map[string]*SSHHostConfig
}

// NewSSHConfig creates a new SSHConfig
func NewSSHConfig() *SSHConfig {
	return &SSHConfig{
		hosts: make(map[string]*SSHHostConfig),
	}
}

// LoadSSHConfig loads SSH configuration from ~/.ssh/config
func LoadSSHConfig() (*SSHConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "get user home directory")
	}

	configPath := filepath.Join(homeDir, ".ssh", "config")
	return LoadSSHConfigFromFile(configPath)
}

// LoadSSHConfigFromFile loads SSH configuration from specified file
func LoadSSHConfigFromFile(path string) (*SSHConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return NewSSHConfig(), nil
		}
		return nil, errors.Wrapf(err, "open SSH config file: %s", path)
	}
	defer file.Close()

	config := NewSSHConfig()
	scanner := bufio.NewScanner(file)

	var currentHost *SSHHostConfig
	var currentHostPatterns []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split into key and value
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")

		// Remove quotes if present
		value = strings.Trim(value, "\"'")

		switch key {
		case "host":
			// Save previous host config
			if currentHost != nil {
				for _, pattern := range currentHostPatterns {
					config.hosts[pattern] = currentHost
				}
			}

			// Start new host config
			currentHostPatterns = parts[1:]
			currentHost = &SSHHostConfig{
				Host:                currentHostPatterns[0],
				Port:                22, // default SSH port
				IdentityFiles:       []string{},
				PreferredAuth:       []string{},
				SendEnv:             []string{},
				SetEnv:              make(map[string]string),
				LocalForward:        []string{},
				RemoteForward:       []string{},
				DynamicForward:      []string{},
				ServerAliveInterval: 0,
				ServerAliveCountMax: 3,
				StrictHostKeyCheck:  "ask",
			}

		case "hostname":
			if currentHost != nil {
				currentHost.HostName = value
			}

		case "user":
			if currentHost != nil {
				currentHost.User = value
			}

		case "port":
			if currentHost != nil {
				port := parseIntDefault(value, 22)
				currentHost.Port = port
			}

		case "identityfile":
			if currentHost != nil {
				// Expand ~ to home directory
				if strings.HasPrefix(value, "~/") {
					homeDir, _ := os.UserHomeDir()
					value = filepath.Join(homeDir, value[2:])
				}
				currentHost.IdentityFiles = append(currentHost.IdentityFiles, value)
				if currentHost.IdentityFile == "" {
					currentHost.IdentityFile = value
				}
			}

		case "preferredauthentications":
			if currentHost != nil {
				auths := strings.Split(value, ",")
				for i, auth := range auths {
					auths[i] = strings.TrimSpace(auth)
				}
				currentHost.PreferredAuth = auths
			}

		case "proxyjump":
			if currentHost != nil {
				currentHost.ProxyJump = value
			}

		case "proxycommand":
			if currentHost != nil {
				currentHost.ProxyCommand = value
			}

		case "forwardagent":
			if currentHost != nil {
				currentHost.ForwardAgent = strings.ToLower(value) == "yes"
			}

		case "compression":
			if currentHost != nil {
				currentHost.Compression = strings.ToLower(value) == "yes"
			}

		case "serveraliveinterval":
			if currentHost != nil {
				currentHost.ServerAliveInterval = parseIntDefault(value, 0)
			}

		case "serveralivecountmax":
			if currentHost != nil {
				currentHost.ServerAliveCountMax = parseIntDefault(value, 3)
			}

		case "stricthostkeychecking":
			if currentHost != nil {
				currentHost.StrictHostKeyCheck = value
			}

		case "userknownhostsfile":
			if currentHost != nil {
				currentHost.UserKnownHostsFile = value
			}

		case "connecttimeout":
			if currentHost != nil {
				currentHost.ConnectTimeout = parseIntDefault(value, 0)
			}

		case "passwordauthentication":
			if currentHost != nil {
				currentHost.PasswordAuth = strings.ToLower(value) == "yes"
			}

		case "pubkeyauthentication":
			if currentHost != nil {
				currentHost.PubkeyAuth = strings.ToLower(value) == "yes"
			}

		case "kbdinteractiveauthentication":
			if currentHost != nil {
				currentHost.KbdInteractiveAuth = strings.ToLower(value) == "yes"
			}

		case "rekeylimit":
			if currentHost != nil {
				currentHost.RekeyLimit = value
			}

		case "sendenv":
			if currentHost != nil {
				currentHost.SendEnv = append(currentHost.SendEnv, value)
			}

		case "setenv":
			if currentHost != nil {
				envParts := strings.SplitN(value, "=", 2)
				if len(envParts) == 2 {
					currentHost.SetEnv[envParts[0]] = envParts[1]
				}
			}

		case "requesttty":
			if currentHost != nil {
				currentHost.RequestTTY = value
			}

		case "remotecommand":
			if currentHost != nil {
				currentHost.RemoteCommand = value
			}

		case "localforward":
			if currentHost != nil {
				currentHost.LocalForward = append(currentHost.LocalForward, value)
			}

		case "remoteforward":
			if currentHost != nil {
				currentHost.RemoteForward = append(currentHost.RemoteForward, value)
			}

		case "dynamicforward":
			if currentHost != nil {
				currentHost.DynamicForward = append(currentHost.DynamicForward, value)
			}

		case "exitonforwardfailure":
			if currentHost != nil {
				currentHost.ExitOnForwardFailure = strings.ToLower(value) == "yes"
			}

		case "controlmaster":
			if currentHost != nil {
				currentHost.ControlMaster = value
			}

		case "controlpath":
			if currentHost != nil {
				currentHost.ControlPath = value
			}

		case "controlpersist":
			if currentHost != nil {
				currentHost.ControlPersist = value
			}
		}
	}

	// Save last host config
	if currentHost != nil {
		for _, pattern := range currentHostPatterns {
			config.hosts[pattern] = currentHost
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "read SSH config file")
	}

	return config, nil
}

// GetHostConfig returns SSH configuration for the specified host
func (c *SSHConfig) GetHostConfig(host string) *SSHHostConfig {
	// Try exact match first
	if config, ok := c.hosts[host]; ok {
		return config
	}

	// Try wildcard match
	for pattern, config := range c.hosts {
		if matchSSHPattern(pattern, host) {
			return config
		}
	}

	return nil
}

// matchSSHPattern matches host against SSH config pattern
// Supports * (matches zero or more characters) and ? (matches exactly one character)
func matchSSHPattern(pattern, host string) bool {
	// Exact match
	if pattern == host {
		return true
	}

	// * matches everything
	if pattern == "*" {
		return true
	}

	// If no wildcards, return false
	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
		return false
	}

	// Simple implementation for common cases
	// Handle patterns like "*.example.com"
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // Include the dot
		return strings.HasSuffix(host, suffix)
	}

	// Handle patterns like "server*"
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(host, prefix)
	}

	// Handle patterns like "*server"
	if strings.HasPrefix(pattern, "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(host, suffix)
	}

	// For more complex patterns, use a simple matching algorithm
	return simpleWildcardMatch(pattern, host)
}

// simpleWildcardMatch performs simple wildcard matching
func simpleWildcardMatch(pattern, str string) bool {
	pi, si := 0, 0
	starIdx, matchIdx := -1, 0

	for si < len(str) {
		if pi < len(pattern) && (pattern[pi] == str[si] || pattern[pi] == '?') {
			pi++
			si++
		} else if pi < len(pattern) && pattern[pi] == '*' {
			starIdx = pi
			matchIdx = si
			pi++
		} else if starIdx != -1 {
			pi = starIdx + 1
			matchIdx++
			si = matchIdx
		} else {
			return false
		}
	}

	for pi < len(pattern) && pattern[pi] == '*' {
		pi++
	}

	return pi == len(pattern)
}

// SSHConfigToCredentials converts SSH host config to Credentials
func SSHConfigToCredentials(config *SSHHostConfig) *models.Credentials {
	if config == nil {
		return nil
	}

	creds := &models.Credentials{
		User: config.User,
	}

	// Set identity file (private key) if available
	if config.IdentityFile != "" {
		keyData, err := os.ReadFile(config.IdentityFile)
		if err == nil {
			creds.RsaPrivateKey = string(keyData)
		}
	}

	return creds
}

// ParseSSHURL parses various SSH URL formats:
// - ssh:hostname (uses host from SSH config)
// - ssh://hostname
// - ssh://user@hostname
// - ssh://user@hostname:port
// - ssh://user@hostname:port/path
// - ssh://user@hostname/path
// - sftp://hostname
// - user@hostname (SSH shorthand)
// - user@hostname:path (SSH shorthand with path)
// - user@hostname:/path (SSH shorthand with absolute path)
func ParseSSHURL(urlStr string) (protocol, user, hostname, path string, port int, err error) {
	// Default values
	port = 22
	protocol = "ssh"
	path = ""

	originalURL := urlStr

	// Handle ssh: shorthand format (ssh:hostname)
	if strings.HasPrefix(urlStr, "ssh:") && !strings.HasPrefix(urlStr, "ssh://") {
		hostname = strings.TrimPrefix(urlStr, "ssh:")
		return protocol, user, hostname, path, port, nil
	}

	// Handle sftp: shorthand format (sftp:hostname)
	if strings.HasPrefix(urlStr, "sftp:") && !strings.HasPrefix(urlStr, "sftp://") {
		protocol = "sftp"
		hostname = strings.TrimPrefix(urlStr, "sftp:")
		return protocol, user, hostname, path, port, nil
	}

	// Handle standard URL format (ssh:// or sftp://)
	if strings.HasPrefix(urlStr, "ssh://") {
		urlStr = strings.TrimPrefix(urlStr, "ssh://")
	} else if strings.HasPrefix(urlStr, "sftp://") {
		protocol = "sftp"
		urlStr = strings.TrimPrefix(urlStr, "sftp://")
	}

	// Handle SSH shorthand format: user@host or user@host:path
	if !strings.HasPrefix(originalURL, "ssh://") && !strings.HasPrefix(originalURL, "sftp://") {
		if strings.Contains(urlStr, "@") {
			// This is SSH shorthand format
			// Parse user@hostname:path or user@hostname:/path or user@hostname
			parts := strings.SplitN(urlStr, "@", 2)
			user = parts[0]
			remainder := parts[1]

			// Check for path separator
			if strings.Contains(remainder, ":") {
				// Split hostname and path
				colonIdx := strings.Index(remainder, ":")
				hostname = remainder[:colonIdx]
				path = remainder[colonIdx+1:]

				// If path starts with a digit, it might be a port number
				// Check if it's followed by / or if it's purely numeric up to the next /
				if len(path) > 0 && path[0] >= '0' && path[0] <= '9' {
					// Find the first non-digit character
					portEndIdx := 0
					for portEndIdx < len(path) && path[portEndIdx] >= '0' && path[portEndIdx] <= '9' {
						portEndIdx++
					}

					// If we have only digits, or digits followed by /, it's a port
					if portEndIdx > 0 && (portEndIdx == len(path) || path[portEndIdx] == '/') {
						portStr := path[:portEndIdx]
						port = parseIntDefault(portStr, 22)
						if portEndIdx < len(path) {
							path = path[portEndIdx:] // Keep the / and path
						} else {
							path = ""
						}
					}
				}
			} else {
				hostname = remainder
			}

			if hostname == "" {
				err = errors.New("hostname cannot be empty")
				return
			}

			return protocol, user, hostname, path, port, nil
		}
	}

	// Parse standard URL format: [user@]hostname[:port][/path]
	var hostPort string
	if strings.Contains(urlStr, "@") {
		parts := strings.SplitN(urlStr, "@", 2)
		user = parts[0]
		hostPort = parts[1]
	} else {
		hostPort = urlStr
	}

	// Parse path
	if strings.Contains(hostPort, "/") {
		slashIdx := strings.Index(hostPort, "/")
		path = hostPort[slashIdx:]
		hostPort = hostPort[:slashIdx]
	}

	// Parse hostname:port
	if strings.Contains(hostPort, ":") {
		parts := strings.SplitN(hostPort, ":", 2)
		hostname = parts[0]
		port = parseIntDefault(parts[1], 22)
	} else {
		hostname = hostPort
	}

	if hostname == "" {
		err = errors.New("hostname cannot be empty")
		return
	}

	return protocol, user, hostname, path, port, nil
}

// NewCredentialsFromSSH creates Credentials from SSH URL using ~/.ssh/config
func NewCredentialsFromSSH(sshURL string) (*models.Credentials, string, int, string, error) {
	// Parse SSH URL
	_, user, hostname, path, port, err := ParseSSHURL(sshURL)
	if err != nil {
		return nil, "", 0, "", errors.Wrap(err, "parse SSH URL")
	}

	// Load SSH config
	sshConfig, err := LoadSSHConfig()
	if err != nil {
		// If SSH config cannot be loaded, return basic credentials
		return &models.Credentials{User: user}, hostname, port, path, nil
	}

	// Get host configuration
	hostConfig := sshConfig.GetHostConfig(hostname)
	if hostConfig == nil {
		// No config found, return basic credentials
		return &models.Credentials{User: user}, hostname, port, path, nil
	}

	// Create credentials from SSH config
	creds := SSHConfigToCredentials(hostConfig)

	// Override with explicit values from URL
	if user != "" {
		creds.User = user
	}

	// Use hostname from config if available
	actualHostname := hostname
	if hostConfig.HostName != "" {
		actualHostname = hostConfig.HostName
	}

	// Use port from config if not explicitly specified in URL
	actualPort := port
	if hostConfig.Port != 22 {
		actualPort = hostConfig.Port
	}

	return creds, actualHostname, actualPort, path, nil
}

// BuildSSHURL builds a full SSH URL from credentials and hostname
func BuildSSHURL(protocol, user, hostname string, port int, path string) string {
	if protocol == "" {
		protocol = "ssh"
	}

	url := fmt.Sprintf("%s://", protocol)

	if user != "" {
		url += user + "@"
	}

	url += hostname

	if port != 22 {
		url += fmt.Sprintf(":%d", port)
	}

	if path != "" {
		if !strings.HasPrefix(path, "/") {
			url += "/"
		}
		url += path
	}

	return url
}
