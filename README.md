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

## Installation

```bash
# Clone the repository (if you're using git)
git clone https://github.com/yourusername/goflood.git
cd goflood

# Build the executable
go build -o goflood

# Alternative: Install directly with Go
go install github.com/yourusername/goflood@latest
