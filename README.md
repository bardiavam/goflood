# GoFlood - Proxy-enabled HTTP/HTTPS Flood Testing Tool

GoFlood is a security research tool for testing web infrastructure resilience against distributed denial-of-service (DDoS) attacks. It provides a framework to simulate HTTP/HTTPS flood attacks using proxy servers.

## ⚠️ IMPORTANT DISCLAIMER

This tool is provided **FOR SECURITY RESEARCH PURPOSES ONLY**. Using this tool against any website or service without explicit permission from the owner is:

1. Likely **ILLEGAL** in most jurisdictions
2. A violation of most service providers' terms of service
3. **UNETHICAL** and potentially harmful to legitimate users

**THE AUTHORS TAKE NO RESPONSIBILITY** for any misuse of this software. Users are solely responsible for ensuring they have proper authorization before conducting any security testing.

## Features

- HTTP/HTTPS flooding capability through proxies
- Configurable concurrency level (number of workers)
- Proxy validation before attack
- Automatic proxy rotation
- Detailed statistics reporting
- Graceful shutdown support
- Customizable request parameters
- Automatic proxy grabber from public sources

## Installation

```bash
# Clone the repository (if you're using git)
git clone https://github.com/yourusername/goflood.git
cd goflood

# Build the executable
go build -o goflood

# Alternative: Install directly with Go
go install github.com/yourusername/goflood@latest
```

## Usage

```bash
# Basic usage
./goflood -target http://example.com -proxies proxies.txt -duration 30s

# Advanced usage with all options
./goflood -target https://example.com -proxies proxies.txt -workers 1000 -duration 2m -proxycheck http://google.com -skipcheck -verbose

# Using the proxy grabber
./goflood -target http://example.com -grabproxies -graboutput my_proxies.txt -duration 1m
```

### Required Flags

- `-target`: The target URL to attack (must include http:// or https://)
- `-proxies` or `-grabproxies`: Either specify a file with proxies or use the grabber to fetch them

### Optional Flags

- `-workers`: Number of concurrent workers (default: 500)
- `-duration`: Duration of the attack (default: 60s, format: 30s, 5m, 2h)
- `-proxycheck`: URL to use for proxy checking (default: Google's generate_204 endpoint)
- `-skipcheck`: Skip the proxy validation phase (use all proxies regardless of whether they work)
- `-verbose`: Enable verbose logging during the attack

### Proxy Grabber Options

- `-grabproxies`: Enable automatic proxy grabbing from public sources
- `-graboutput`: Filename to save grabbed proxies (default: proxies.txt)
