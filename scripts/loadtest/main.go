// Command loadtest is a minimal HTTP load generator for fluxbase baselines.
// It uses only the standard library so any environment that can build
// fluxbase can run it — no hey/k6/wrk required.
//
// Usage:
//
//	go run ./scripts/loadtest -url http://localhost:8080 \
//	    -paths /livez,/health -duration 30s -c 8 [-token $TOKEN] [-header k:v]
//
// Capture the server's memory footprint during the run:
//
//	ps -o rss=,vsz=,pid=,cmd= -C fluxbase        # native binary
//	docker stats --no-stream fluxbase            # container
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "base URL of the fluxbase server")
	paths := flag.String("paths", "/livez,/health", "comma-separated paths to request round-robin")
	duration := flag.Duration("duration", 30*time.Second, "how long to generate load")
	concurrency := flag.Int("c", 8, "number of concurrent workers")
	token := flag.String("token", "", "optional bearer token sent as Authorization header")
	header := flag.String("header", "", "optional extra header, format \"Name: value\" (repeat not supported)")
	timeout := flag.Duration("timeout", 10*time.Second, "per-request timeout")
	flag.Parse()

	urls := strings.Split(*paths, ",")
	for i, p := range urls {
		urls[i] = strings.TrimRight(*baseURL, "/") + strings.TrimSpace(p)
	}

	client := &http.Client{Timeout: *timeout}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = *concurrency
	client.Transport = transport

	var (
		total     atomic.Int64
		errors    atomic.Int64
		statusCnt sync.Map // status code -> *atomic.Int64
		latencies = make([][]time.Duration, *concurrency)
		done      = make(chan struct{})
	)

	deadline := time.Now().Add(*duration)
	start := time.Now()

	var wg sync.WaitGroup
	for w := 0; w < *concurrency; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			lat := latencies[w]
			for i := 0; time.Now().Before(deadline); i++ {
				url := urls[i%len(urls)]

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					errors.Add(1)
					continue
				}
				if *token != "" {
					req.Header.Set("Authorization", "Bearer "+*token)
				}
				if *header != "" {
					if k, v, ok := strings.Cut(*header, ":"); ok {
						req.Header.Set(strings.TrimSpace(k), strings.TrimSpace(v))
					}
				}

				t0 := time.Now()
				resp, err := client.Do(req)
				lat = append(lat, time.Since(t0))
				if err != nil {
					errors.Add(1)
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				total.Add(1)
				cnt, _ := statusCnt.LoadOrStore(fmt.Sprint(resp.StatusCode), new(atomic.Int64))
				cnt.(*atomic.Int64).Add(1)
			}
			latencies[w] = lat
		}(w)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	// Progress line every 5 seconds until done.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			report(start, *concurrency, &total, &errors, &statusCnt, latencies)
			return
		case <-ticker.C:
			fmt.Printf("  %4ds  requests=%d errors=%d\n",
				int(time.Since(start).Seconds()), total.Load(), errors.Load())
		}
	}
}

func report(start time.Time, concurrency int, total, errors *atomic.Int64,
	statusCnt *sync.Map, latencies [][]time.Duration,
) {
	elapsed := time.Since(start)

	var all []time.Duration
	for _, lat := range latencies {
		all = append(all, lat...)
	}
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })

	pct := func(p float64) time.Duration {
		if len(all) == 0 {
			return 0
		}
		idx := int(float64(len(all)-1) * p)
		return all[idx]
	}

	fmt.Printf("\n=== Baseline results (%s, concurrency %d) ===\n", elapsed.Round(time.Millisecond), concurrency)
	fmt.Printf("requests:  %d\n", total.Load())
	fmt.Printf("errors:    %d\n", errors.Load())
	fmt.Printf("throughput: %.1f req/s\n", float64(total.Load())/elapsed.Seconds())
	if len(all) > 0 {
		fmt.Printf("latency:   p50=%s p90=%s p99=%s max=%s\n",
			pct(0.50).Round(10*time.Microsecond), pct(0.90).Round(10*time.Microsecond),
			pct(0.99).Round(10*time.Microsecond), all[len(all)-1].Round(10*time.Microsecond))
	}
	statusCnt.Range(func(k, v any) bool {
		fmt.Printf("status %s: %d\n", k, v.(*atomic.Int64).Load())
		return true
	})
	fmt.Println("\nRecord the server's RSS during the run too (see header comment).")
	os.Exit(0)
}
