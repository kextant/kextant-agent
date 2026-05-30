package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	RedactionPlain    = "plain"
	RedactionHashed   = "hashed"
	RedactionOmitted  = "omitted"
	MetadataAll       = "all"
	MetadataAllowlist = "allowlist"
	MetadataNone      = "none"
)

var (
	DefaultLabelAllowlist = []string{
		"app",
		"k8s-app",
		"name",
		"version",
		"app.kubernetes.io/name",
		"app.kubernetes.io/instance",
		"app.kubernetes.io/version",
		"app.kubernetes.io/component",
		"app.kubernetes.io/part-of",
		"app.kubernetes.io/managed-by",
		"helm.sh/chart",
	}

	DefaultAnnotationAllowlist = []string{
		"meta.helm.sh/release-name",
		"meta.helm.sh/release-namespace",
	}
)

type Config struct {
	// General settings
	ClusterName    string
	ClusterID      string
	ReportSchedule string
	Timezone       string

	// Health checks
	ChecksResourcesEnabled       bool
	ChecksProbesEnabled          bool
	ChecksReplicasEnabled        bool
	ChecksImagesEnabled          bool
	ChecksSecurityContextEnabled bool
	ChecksDeprecatedAPIEnabled   bool
	ChecksNamespaceEnabled       bool

	// Inventory collection
	HelmMetadataEnabled bool

	// Namespaces
	ScanNamespaces    []string
	ExcludeNamespaces []string

	// Exemptions
	ExemptChecks []string

	// Privacy/redaction
	RedactionNamespaceNames      string
	RedactionWorkloadNames       string
	RedactionLabels              string
	RedactionAnnotations         string
	RedactionLabelAllowlist      []string
	RedactionAnnotationAllowlist []string

	// Kextant Cloud upload
	KextantCloudEnabled       bool
	KextantCloudEndpoint      string
	KextantCloudAPIKey        string
	KextantCloudTLSSkipVerify bool
	KextantCloudUploadRetries int
	KextantCloudMTLSCertFile  string
	KextantCloudMTLSKeyFile   string

	// Delivery targets
	SlackEnabled bool
	EmailEnabled bool
	LogEnabled   bool
	LogMultiline bool

	// Email settings
	EmailRecipients []string
	EmailFrom       string
	EmailSubject    string

	// SMTP configuration
	SMTPHost          string
	SMTPPort          string
	SMTPAuthType      string
	SMTPTLSMode       string
	SMTPTLSSkipVerify bool

	// Report settings
	ReportMinSeverity          string
	ReportIncludePassed        bool
	ReportMaxIssuesPerCategory int

	// Secrets
	SlackWebhookURL string
	SMTPUsername    string
	SMTPPassword    string
}

func Load() (*Config, error) {
	cfg := &Config{
		// General settings
		ClusterName:    getEnv("CLUSTER_NAME", "kubernetes-cluster"),
		ClusterID:      getEnv("CLUSTER_ID", ""),
		ReportSchedule: getEnv("REPORT_SCHEDULE", "0 9 * * 1"),
		Timezone:       getEnv("TIMEZONE", "UTC"),

		// Health checks
		ChecksResourcesEnabled:       getBoolEnv("CHECKS_RESOURCES_ENABLED", true),
		ChecksProbesEnabled:          getBoolEnv("CHECKS_PROBES_ENABLED", true),
		ChecksReplicasEnabled:        getBoolEnv("CHECKS_REPLICAS_ENABLED", true),
		ChecksImagesEnabled:          getBoolEnv("CHECKS_IMAGES_ENABLED", true),
		ChecksSecurityContextEnabled: getBoolEnv("CHECKS_SECURITY_CONTEXT_ENABLED", true),
		ChecksDeprecatedAPIEnabled:   getBoolEnv("CHECKS_DEPRECATED_API_ENABLED", true),
		ChecksNamespaceEnabled:       getBoolEnv("CHECKS_NAMESPACE_ENABLED", true),

		// Inventory collection
		HelmMetadataEnabled: getBoolEnv("HELM_METADATA_ENABLED", false),

		// Namespaces
		ScanNamespaces:    getSliceEnv("SCAN_NAMESPACES"),
		ExcludeNamespaces: getSliceEnv("EXCLUDE_NAMESPACES"),

		// Exemptions
		ExemptChecks: getSliceEnv("EXEMPT_CHECKS"),

		// Privacy/redaction
		RedactionNamespaceNames:      getEnv("REDACTION_NAMESPACE_NAMES", RedactionPlain),
		RedactionWorkloadNames:       getEnv("REDACTION_WORKLOAD_NAMES", RedactionPlain),
		RedactionLabels:              getEnv("REDACTION_LABELS", MetadataAllowlist),
		RedactionAnnotations:         getEnv("REDACTION_ANNOTATIONS", MetadataAllowlist),
		RedactionLabelAllowlist:      getSliceEnvDefault("REDACTION_LABEL_ALLOWLIST", DefaultLabelAllowlist),
		RedactionAnnotationAllowlist: getSliceEnvDefault("REDACTION_ANNOTATION_ALLOWLIST", DefaultAnnotationAllowlist),

		// Kextant Cloud upload
		KextantCloudEnabled:       getBoolEnv("KEXTANT_CLOUD_ENABLED", false),
		KextantCloudEndpoint:      getEnv("KEXTANT_CLOUD_ENDPOINT", "https://api.kextant.com"),
		KextantCloudAPIKey:        getEnv("KEXTANT_CLOUD_API_KEY", ""),
		KextantCloudTLSSkipVerify: getBoolEnv("KEXTANT_CLOUD_TLS_SKIP_VERIFY", false),
		KextantCloudUploadRetries: getIntEnv("KEXTANT_CLOUD_UPLOAD_RETRIES", 3),
		KextantCloudMTLSCertFile:  getEnv("KEXTANT_CLOUD_MTLS_CERT_FILE", ""),
		KextantCloudMTLSKeyFile:   getEnv("KEXTANT_CLOUD_MTLS_KEY_FILE", ""),

		// Delivery targets
		SlackEnabled: getBoolEnv("SLACK_ENABLED", false),
		EmailEnabled: getBoolEnv("EMAIL_ENABLED", false),
		LogEnabled:   getBoolEnv("LOG_ENABLED", true),
		LogMultiline: getBoolEnv("LOG_MULTILINE", true),

		// Email settings
		EmailRecipients: getSliceEnv("EMAIL_RECIPIENTS"),
		EmailFrom:       getEnv("EMAIL_FROM", "kextant@example.com"),
		EmailSubject:    getEnv("EMAIL_SUBJECT", "Kextant Health Report: {{.ClusterName}} - {{.Date}}"),

		// SMTP configuration
		SMTPHost:          getEnv("SMTP_HOST", ""),
		SMTPPort:          getEnv("SMTP_PORT", "587"),
		SMTPAuthType:      getEnv("SMTP_AUTH_TYPE", "plain"),
		SMTPTLSMode:       getEnv("SMTP_TLS_MODE", "starttls"),
		SMTPTLSSkipVerify: getBoolEnv("SMTP_TLS_SKIP_VERIFY", false),

		// Report settings
		ReportMinSeverity:          getEnv("REPORT_MIN_SEVERITY", "warning"),
		ReportIncludePassed:        getBoolEnv("REPORT_INCLUDE_PASSED", false),
		ReportMaxIssuesPerCategory: getIntEnv("REPORT_MAX_ISSUES_PER_CATEGORY", 10),

		// Secrets
		SlackWebhookURL: getEnv("SLACK_WEBHOOK_URL", ""),
		SMTPUsername:    getEnv("SMTP_USERNAME", ""),
		SMTPPassword:    getEnv("SMTP_PASSWORD", ""),
	}

	// Set default exclude namespaces if not specified.
	if len(cfg.ExcludeNamespaces) == 0 {
		cfg.ExcludeNamespaces = []string{"kube-system", "kube-public", "kube-node-lease", "kextant"}
	}

	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if c.SlackEnabled && c.SlackWebhookURL == "" {
		return fmt.Errorf("SLACK_ENABLED is true but SLACK_WEBHOOK_URL is not set")
	}

	if c.EmailEnabled {
		if len(c.EmailRecipients) == 0 {
			return fmt.Errorf("EMAIL_ENABLED is true but EMAIL_RECIPIENTS is not set")
		}
		if c.SMTPHost == "" {
			return fmt.Errorf("EMAIL_ENABLED is true but SMTP_HOST is not set")
		}
	}

	if err := validateRedactionMode("REDACTION_NAMESPACE_NAMES", c.RedactionNamespaceNames); err != nil {
		return err
	}
	if err := validateRedactionMode("REDACTION_WORKLOAD_NAMES", c.RedactionWorkloadNames); err != nil {
		return err
	}
	if err := validateMetadataMode("REDACTION_LABELS", c.RedactionLabels); err != nil {
		return err
	}
	if err := validateMetadataMode("REDACTION_ANNOTATIONS", c.RedactionAnnotations); err != nil {
		return err
	}

	if c.KextantCloudEnabled {
		if err := c.ValidateCloudUpload(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) ValidateCloudUpload() error {
	if c.ClusterID == "" {
		return fmt.Errorf("CLUSTER_ID is required for Kextant Cloud upload")
	}
	if c.KextantCloudEndpoint == "" {
		return fmt.Errorf("KEXTANT_CLOUD_ENDPOINT is required for Kextant Cloud upload")
	}
	parsed, err := url.Parse(c.KextantCloudEndpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("KEXTANT_CLOUD_ENDPOINT must be a valid URL")
	}
	if parsed.Scheme != "https" && !c.KextantCloudTLSSkipVerify {
		return fmt.Errorf("KEXTANT_CLOUD_ENDPOINT must use https unless KEXTANT_CLOUD_TLS_SKIP_VERIFY is true for development")
	}
	if c.KextantCloudAPIKey == "" {
		return fmt.Errorf("KEXTANT_CLOUD_API_KEY is required for Kextant Cloud upload")
	}
	if (c.KextantCloudMTLSCertFile == "") != (c.KextantCloudMTLSKeyFile == "") {
		return fmt.Errorf("KEXTANT_CLOUD_MTLS_CERT_FILE and KEXTANT_CLOUD_MTLS_KEY_FILE must be provided together")
	}
	if c.KextantCloudUploadRetries < 0 {
		return fmt.Errorf("KEXTANT_CLOUD_UPLOAD_RETRIES must be >= 0")
	}
	return nil
}

// ValidateLogDelivery checks if LOG_ENABLED is compatible with the logging configuration.
// Returns a warning message if incompatible, empty string otherwise.
func (c *Config) ValidateLogDelivery() string {
	if !c.LogEnabled {
		return ""
	}

	// Check if LOG_LEVEL environment variable is set to something other than INFO or DEBUG.
	logLevel := strings.ToUpper(getEnv("LOG_LEVEL", "INFO"))

	// If log level is WARN or ERROR, log delivery won't work.
	if logLevel == "WARN" || logLevel == "WARNING" || logLevel == "ERROR" {
		return fmt.Sprintf("LOG_ENABLED is true but LOG_LEVEL is set to %s (must be INFO or DEBUG for log delivery to work)", logLevel)
	}

	return ""
}

func validateRedactionMode(name, value string) error {
	switch value {
	case RedactionPlain, RedactionHashed, RedactionOmitted:
		return nil
	default:
		return fmt.Errorf("%s must be one of %q, %q, or %q", name, RedactionPlain, RedactionHashed, RedactionOmitted)
	}
}

func validateMetadataMode(name, value string) error {
	switch value {
	case MetadataAll, MetadataAllowlist, MetadataNone:
		return nil
	default:
		return fmt.Errorf("%s must be one of %q, %q, or %q", name, MetadataAll, MetadataAllowlist, MetadataNone)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return b
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return i
	}
	return defaultValue
}

func getSliceEnv(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	return splitCSV(value)
}

func getSliceEnvDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return splitCSV(value)
	}
	result := make([]string, len(defaultValue))
	copy(result, defaultValue)
	return result
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
