package main

import (
        "bufio"
        "context"
        "fmt"
        "io"
        "log"
        "net/http"
        "net/url"
        "os"
        "strings"
        "sync"
        "time"
)

const (
        // Maximum number of concurrent proxy checks
        maxConcurrentChecks = 200
        // Default timeout for grabber operations
        grabberTimeout = 30 * time.Second
        // Default file to save proxies
        defaultProxyFile = "proxies.txt"
)

// ProxySources defines URLs where we can find public proxy lists
var ProxySources = []string{
        "https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
        "https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks4.txt",
        "https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt",
        "https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/http.txt",
        "https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/https.txt",
        "https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks4.txt",
        "https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks5.txt",
        "https://raw.githubusercontent.com/hookzof/socks5_list/master/proxy.txt",
        "https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-raw.txt",
        "https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt",
        "https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/socks4.txt",
        "https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/socks5.txt",
}

// GrabProxies fetches proxies from public sources and saves working ones to a file
func GrabProxies(outputFile string, checkURL string, verbose bool) (int, error) {
        if outputFile == "" {
                outputFile = defaultProxyFile
        }

        // Make sure the checkURL is valid
        if !strings.HasPrefix(checkURL, "http://") && !strings.HasPrefix(checkURL, "https://") {
                return 0, fmt.Errorf("check URL must start with http:// or https://")
        }

        log.Println("Starting proxy grabber...")
        
        // Create a context with timeout for the entire operation
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
        defer cancel()

        // Collect all proxies from sources
        allProxies := make(map[string]bool) // Using map to eliminate duplicates
        for _, source := range ProxySources {
                if verbose {
                        log.Printf("Fetching proxies from %s...", source)
                }
                
                // Create an HTTP client with timeout
                client := &http.Client{
                        Timeout: grabberTimeout,
                }
                
                // Create a request with context
                req, err := http.NewRequestWithContext(ctx, "GET", source, nil)
                if err != nil {
                        log.Printf("Warning: Failed to create request for %s: %v", source, err)
                        continue
                }
                
                // Add a common user agent to avoid rejections
                req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
                
                // Send the request
                resp, err := client.Do(req)
                if err != nil {
                        log.Printf("Warning: Failed to fetch from %s: %v", source, err)
                        continue
                }
                
                // Always close response bodies
                defer resp.Body.Close()
                
                // Check if the response was successful
                if resp.StatusCode != http.StatusOK {
                        log.Printf("Warning: Source %s returned status %d", source, resp.StatusCode)
                        continue
                }
                
                // Read the response body
                body, err := io.ReadAll(resp.Body)
                if err != nil {
                        log.Printf("Warning: Failed to read response from %s: %v", source, err)
                        continue
                }
                
                // Parse the proxies
                scanner := bufio.NewScanner(strings.NewReader(string(body)))
                for scanner.Scan() {
                        proxy := strings.TrimSpace(scanner.Text())
                        if proxy != "" {
                                // Make sure we have scheme for proper parsing
                                if !strings.Contains(proxy, "://") {
                                        // Check if we can identify scheme from source URL
                                        if strings.Contains(source, "http.txt") {
                                                proxy = "http://" + proxy
                                        } else if strings.Contains(source, "https.txt") {
                                                proxy = "https://" + proxy
                                        } else if strings.Contains(source, "socks4.txt") {
                                                proxy = "socks4://" + proxy
                                        } else if strings.Contains(source, "socks5.txt") {
                                                proxy = "socks5://" + proxy
                                        } else {
                                                // Default to http if unknown
                                                proxy = "http://" + proxy
                                        }
                                }
                                
                                // Add to our map to avoid duplicates
                                allProxies[proxy] = true
                        }
                }
                
                if scanner.Err() != nil {
                        log.Printf("Warning: Error scanning proxies from %s: %v", source, scanner.Err())
                }
        }
        
        // Create a slice of all unique proxies
        var uniqueProxies []string
        for proxy := range allProxies {
                uniqueProxies = append(uniqueProxies, proxy)
        }
        
        log.Printf("Found %d unique proxies from all sources", len(uniqueProxies))
        
        // Check all the proxies concurrently
        workingProxies := checkProxiesForGrabber(ctx, uniqueProxies, checkURL, verbose)
        
        // Save working proxies to the output file
        workingCount := len(workingProxies)
        if workingCount > 0 {
                if err := saveProxiesToFile(workingProxies, outputFile); err != nil {
                        return 0, fmt.Errorf("failed to save proxies to file: %v", err)
                }
                log.Printf("Saved %d working proxies to %s", workingCount, outputFile)
        } else {
                log.Println("No working proxies found")
        }
        
        return workingCount, nil
}

// checkProxiesForGrabber is similar to the main proxy checker but optimized for the grabber
func checkProxiesForGrabber(ctx context.Context, proxies []string, checkURL string, verbose bool) []string {
        var workingProxies []string
        var wg sync.WaitGroup
        var mutex sync.Mutex // To safely append to workingProxies
        
        // Create a semaphore to limit concurrent checks
        sem := make(chan struct{}, maxConcurrentChecks)
        
        // Process counter for display
        var processed int
        var processedMutex sync.Mutex
        total := len(proxies)
        
        // Progress reporting goroutine
        if verbose {
                go func() {
                        ticker := time.NewTicker(3 * time.Second)
                        defer ticker.Stop()
                        
                        for {
                                select {
                                case <-ctx.Done():
                                        return
                                case <-ticker.C:
                                        processedMutex.Lock()
                                        log.Printf("Checking progress: %d/%d (%.1f%%)", 
                                                processed, total, float64(processed)*100/float64(total))
                                        processedMutex.Unlock()
                                }
                        }
                }()
        }
        
        // Check each proxy
        for _, proxyStr := range proxies {
                // Skip if context is done
                if ctx.Err() != nil {
                        break
                }
                
                wg.Add(1)
                sem <- struct{}{} // Acquire a slot
                
                go func(proxyStr string) {
                        defer wg.Done()
                        defer func() { <-sem }() // Release the slot
                        defer func() {
                                processedMutex.Lock()
                                processed++
                                processedMutex.Unlock()
                        }()
                        
                        // Parse the proxy URL
                        proxyURL, err := url.Parse(proxyStr)
                        if err != nil {
                                if verbose {
                                        log.Printf("Invalid proxy URL: %s - %v", proxyStr, err)
                                }
                                return
                        }
                        
                        // Check if this proxy works
                        checkCtx, cancel := context.WithTimeout(ctx, proxyCheckTimeout)
                        defer cancel()
                        
                        if isProxyWorking(checkCtx, proxyURL, checkURL) {
                                mutex.Lock()
                                workingProxies = append(workingProxies, proxyStr)
                                if verbose {
                                        log.Printf("Found working proxy: %s", proxyStr)
                                }
                                mutex.Unlock()
                        }
                }(proxyStr)
        }
        
        wg.Wait()
        return workingProxies
}

// saveProxiesToFile saves a list of proxy URLs to a file
func saveProxiesToFile(proxies []string, filePath string) error {
        file, err := os.Create(filePath)
        if err != nil {
                return err
        }
        defer file.Close()
        
        writer := bufio.NewWriter(file)
        for _, proxy := range proxies {
                fmt.Fprintln(writer, proxy)
        }
        
        return writer.Flush()
}

// getProxyTypeCounts returns a breakdown of proxy types by scheme
func getProxyTypeCounts(proxies []string) map[string]int {
        counts := make(map[string]int)
        
        for _, proxyStr := range proxies {
                proxyURL, err := url.Parse(proxyStr)
                if err != nil {
                        continue
                }
                
                scheme := proxyURL.Scheme
                counts[scheme]++
        }
        
        return counts
}