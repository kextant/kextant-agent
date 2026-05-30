package inventory

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/kextant/kextant-agent/internal/config"
)

type redactor struct {
	cfg *config.Config
}

func newRedactor(cfg *config.Config) redactor {
	return redactor{cfg: cfg}
}

func (r redactor) namespace(value string) string {
	return redactName(value, r.cfg.RedactionNamespaceNames)
}

func (r redactor) workload(value string) string {
	return redactName(value, r.cfg.RedactionWorkloadNames)
}

func (r redactor) labels(values map[string]string) map[string]string {
	return redactMetadata(values, r.cfg.RedactionLabels, r.cfg.RedactionLabelAllowlist)
}

func (r redactor) annotations(values map[string]string) map[string]string {
	return redactMetadata(values, r.cfg.RedactionAnnotations, r.cfg.RedactionAnnotationAllowlist)
}

func redactName(value, mode string) string {
	switch mode {
	case config.RedactionOmitted:
		return ""
	case config.RedactionHashed:
		return stableHash(value)
	default:
		return value
	}
}

func redactMetadata(values map[string]string, mode string, allowlist []string) map[string]string {
	if len(values) == 0 || mode == config.MetadataNone {
		return nil
	}
	result := map[string]string{}
	switch mode {
	case config.MetadataAll:
		for key, value := range values {
			result[key] = value
		}
	case config.MetadataAllowlist:
		allowed := map[string]struct{}{}
		for _, key := range allowlist {
			allowed[key] = struct{}{}
		}
		for key, value := range values {
			if _, ok := allowed[key]; ok {
				result[key] = value
			}
		}
	}
	if len(result) == 0 {
		return nil
	}
	return sortedMap(result)
}

func stableHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])[:16]
}

func sortedMap(values map[string]string) map[string]string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(map[string]string, len(values))
	for _, key := range keys {
		result[key] = values[key]
	}
	return result
}
