package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/example/conncheck/internal/checker"
	"github.com/example/conncheck/internal/config"
	"github.com/redis/go-redis/v9/logging"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "YAML configuration file")
	concurrency := flag.Int("concurrency", 0, "number of parallel checks; 0 uses one worker per target")
	output := flag.String("output", "table", "output format: table or json")
	flag.Parse()

	logging.Disable()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	results := checker.Run(ctx, cfg, *concurrency)
	if err := printResults(*output, results); err != nil {
		fmt.Fprintf(os.Stderr, "output error: %v\n", err)
		os.Exit(2)
	}

	failed := 0
	for _, result := range results {
		if !result.Success {
			failed++
		}
	}
	if failed > 0 {
		os.Exit(1)
	}
}

func printResults(format string, results []checker.Result) error {
	switch format {
	case "table":
		writer := tabwriter.NewWriter(os.Stdout, 0, 4, 3, ' ', 0)
		fmt.Fprintln(writer, "STATUS\tKIND\tNAME\tTARGET\tLATENCY\tERROR")
		for _, result := range results {
			status := "FAIL"
			if result.Success {
				status = "OK"
			}
			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n", status, result.Kind, result.Name, result.Target, result.Latency.Round(time.Millisecond), result.Error)
		}
		return writer.Flush()
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(results)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}
