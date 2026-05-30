package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kextant/kextant-agent/internal/cloud"
	"github.com/kextant/kextant-agent/internal/config"
	"github.com/kextant/kextant-agent/internal/delivery"
	"github.com/kextant/kextant-agent/internal/health"
	"github.com/kextant/kextant-agent/internal/inventory"
	"github.com/kextant/kextant-agent/internal/report"
	"github.com/kextant/kextant-agent/internal/scanner"
	"github.com/kextant/kextant-agent/pkg/types"
	"github.com/robfig/cron/v3"
)

var (
	version   = "0.1.0"
	commit    = "unknown"
	buildDate = "unknown"
)

type scanOptions struct {
	sendReport     bool
	uploadManifest bool
	printManifest  bool
	manifestOutput string
}

type inventoryOptions struct {
	upload bool
	output string
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "version":
		if hasFlag(args, "--json") {
			printVersionJSON()
			return
		}
		fmt.Printf("kextant-agent version %s\n", version)
	case "validate":
		if err := validateConfig(); err != nil {
			logger.Error("configuration validation failed", "error", err)
			os.Exit(1)
		}
		logger.Info("configuration is valid")
	case "scan":
		options := scanOptions{
			sendReport:     hasFlag(args, "--send"),
			uploadManifest: hasFlag(args, "--upload"),
			printManifest:  hasFlag(args, "--manifest"),
			manifestOutput: flagValue(args, "--manifest-output"),
		}
		if err := runScan(logger, options); err != nil {
			logger.Error("scan failed", "error", err)
			os.Exit(1)
		}
	case "inventory":
		options := inventoryOptions{
			upload: hasFlag(args, "--upload"),
			output: flagValue(args, "--output"),
		}
		if err := runInventory(logger, options); err != nil {
			logger.Error("inventory failed", "error", err)
			os.Exit(1)
		}
	case "serve":
		if err := runServer(logger); err != nil {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: kextant-agent <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  scan                         Run a scan immediately and print report to stdout")
	fmt.Println("  scan --send                  Run a scan and send report to configured destinations")
	fmt.Println("  scan --upload                Run a scan and upload inventory manifest to Kextant Cloud")
	fmt.Println("  scan --manifest              Run a scan and print inventory manifest JSON")
	fmt.Println("  scan --manifest-output FILE  Run a scan and write inventory manifest JSON to FILE")
	fmt.Println("  inventory                    Print inventory manifest JSON to stdout")
	fmt.Println("  inventory --output FILE      Write inventory manifest JSON to FILE")
	fmt.Println("  inventory --upload           Upload inventory manifest to Kextant Cloud")
	fmt.Println("  validate                     Validate configuration")
	fmt.Println("  serve                        Start the scheduled agent (normal operation)")
	fmt.Println("  version [--json]             Print version")
}

func printVersionJSON() {
	payload := map[string]string{
		"version":                 version,
		"commit":                  commit,
		"build_date":              buildDate,
		"manifest_schema_version": types.ManifestSchemaVersion,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(payload)
}

func validateConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Check for log delivery compatibility warning.
	if warning := cfg.ValidateLogDelivery(); warning != "" {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", warning)
	}

	return nil
}

func runScan(logger *slog.Logger, options scanOptions) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	scan, err := scanner.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	ctx := context.Background()
	findings, err := scan.Scan(ctx)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	rep := report.Generate(cfg.ClusterName, findings)

	// Print summary to stdout.
	fmt.Printf("Scan completed: Score %d/100 (%s)\n", rep.Score, rep.Grade)
	fmt.Printf("Critical: %d, Warnings: %d, Info: %d\n", rep.CriticalCount, rep.WarningCount, rep.InfoCount)

	if options.sendReport {
		if err := sendReportToTargets(cfg, rep, logger); err != nil {
			return fmt.Errorf("failed to send report: %w", err)
		}
	} else {
		plainText := report.GeneratePlainText(rep)
		fmt.Println()
		fmt.Println(plainText)
	}

	if options.printManifest || options.uploadManifest || options.manifestOutput != "" {
		manifest, err := collectInventory(ctx, cfg, logger)
		if err != nil {
			return err
		}
		if options.manifestOutput != "" {
			if err := writeManifest(options.manifestOutput, manifest); err != nil {
				return err
			}
		}
		if options.printManifest {
			if err := printManifest(manifest); err != nil {
				return err
			}
		}
		if options.uploadManifest {
			if err := uploadManifest(ctx, cfg, manifest, logger); err != nil {
				return err
			}
		}
	}

	return nil
}

func runInventory(logger *slog.Logger, options inventoryOptions) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()
	manifest, err := collectInventory(ctx, cfg, logger)
	if err != nil {
		return err
	}

	if options.output != "" {
		if err := writeManifest(options.output, manifest); err != nil {
			return err
		}
	}
	if options.upload {
		if err := uploadManifest(ctx, cfg, manifest, logger); err != nil {
			return err
		}
	}
	if options.output == "" && !options.upload {
		return printManifest(manifest)
	}

	return nil
}

func runServer(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check for log delivery compatibility warning.
	if warning := cfg.ValidateLogDelivery(); warning != "" {
		logger.Warn(warning)
	}

	logger.Info("starting kextant agent",
		"version", version,
		"cluster", cfg.ClusterName,
		"schedule", cfg.ReportSchedule)

	scan, err := scanner.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	// Start health check server.
	healthSrv := health.NewServer(scan, logger)
	go func() {
		if err := healthSrv.Start(":8080"); err != nil {
			logger.Error("health server failed", "error", err)
		}
	}()

	// Setup cron scheduler.
	c := cron.New()

	_, err = c.AddFunc(cfg.ReportSchedule, func() {
		logger.Info("running scheduled scan")
		if err := runScheduledScan(cfg, scan, logger); err != nil {
			logger.Error("scheduled scan failed", "error", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to schedule job: %w", err)
	}

	c.Start()
	logger.Info("scheduler started", "schedule", cfg.ReportSchedule)

	// Wait for interrupt.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down")
	c.Stop()
	return nil
}

func runScheduledScan(cfg *config.Config, scan *scanner.Scanner, logger *slog.Logger) error {
	ctx := context.Background()
	findings, err := scan.Scan(ctx)
	if err != nil {
		return err
	}

	rep := report.Generate(cfg.ClusterName, findings)

	logger.Info("scan completed",
		"score", rep.Score,
		"grade", rep.Grade,
		"critical", rep.CriticalCount,
		"warnings", rep.WarningCount,
		"info", rep.InfoCount)

	if err := sendReportToTargets(cfg, rep, logger); err != nil {
		return err
	}

	if cfg.KextantCloudEnabled {
		manifest, err := collectInventory(ctx, cfg, logger)
		if err != nil {
			return err
		}
		if err := uploadManifest(ctx, cfg, manifest, logger); err != nil {
			logger.Error("cloud inventory upload failed", "error", err)
		}
	}

	return nil
}

func sendReportToTargets(cfg *config.Config, rep *types.Report, logger *slog.Logger) error {
	var lastErr error

	// Send to Log.
	if cfg.LogEnabled {
		logDelivery := delivery.NewLogDelivery(logger, cfg.LogMultiline)
		if err := logDelivery.Send(rep); err != nil {
			logger.Error("log delivery failed", "error", err)
			lastErr = err
		}
	}

	// Send to Slack.
	if cfg.SlackEnabled {
		slackDelivery := delivery.NewSlackDelivery(cfg.SlackWebhookURL, logger)
		if err := slackDelivery.Send(rep); err != nil {
			logger.Error("slack delivery failed", "error", err)
			lastErr = err
		}
	}

	// Send via Email.
	if cfg.EmailEnabled {
		emailDelivery := delivery.NewEmailDelivery(cfg, logger)
		if err := emailDelivery.Send(rep); err != nil {
			logger.Error("email delivery failed", "error", err)
			lastErr = err
		}
	}

	if !cfg.LogEnabled && !cfg.SlackEnabled && !cfg.EmailEnabled {
		logger.Warn("no delivery targets enabled")
	}

	return lastErr
}

func collectInventory(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*types.Manifest, error) {
	collector, err := inventory.New(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory collector: %w", err)
	}
	manifest, err := collector.Collect(ctx, version, commit)
	if err != nil {
		return nil, fmt.Errorf("inventory collection failed: %w", err)
	}
	return manifest, nil
}

func printManifest(manifest *types.Manifest) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}

func writeManifest(path string, manifest *types.Manifest) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create manifest output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		return fmt.Errorf("write manifest output file: %w", err)
	}
	return nil
}

func uploadManifest(ctx context.Context, cfg *config.Config, manifest *types.Manifest, logger *slog.Logger) error {
	client, err := cloud.New(cfg, logger)
	if err != nil {
		return err
	}
	response, err := client.UploadManifest(ctx, manifest)
	if err != nil {
		return err
	}
	logger.Info("manifest uploaded to Kextant Cloud",
		"manifest_id", response.ManifestID,
		"report_job_id", response.ReportJobID,
		"report_id", response.ReportID,
		"report_url", response.ReportURL,
		"status", response.Status)
	return nil
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func flagValue(args []string, flag string) string {
	for index, arg := range args {
		if arg == flag && index+1 < len(args) {
			return args[index+1]
		}
	}
	return ""
}
