package scanner

import (
	"context"
	"log/slog"

	"github.com/kextant/kextant-agent/internal/config"
	"github.com/kextant/kextant-agent/pkg/exemptions"
	"github.com/kextant/kextant-agent/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Scanner struct {
	client        kubernetes.Interface
	dynamicClient dynamic.Interface
	config        *config.Config
	logger        *slog.Logger
	exemptions    *exemptions.Resolver
}

func New(cfg *config.Config, logger *slog.Logger) (*Scanner, error) {
	// Try in-cluster config first, fall back to kubeconfig
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig for local development
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)
		k8sConfig, err = kubeconfig.ClientConfig()
		if err != nil {
			return nil, err
		}
	}

	client, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &Scanner{
		client:        client,
		dynamicClient: dynamicClient,
		config:        cfg,
		logger:        logger,
		exemptions:    exemptions.NewResolver(client, cfg.ExemptChecks),
	}, nil
}

func (s *Scanner) Scan(ctx context.Context) ([]types.Finding, error) {
	s.logger.Info("starting scan", "cluster", s.config.ClusterName)

	// Clear exemption cache at start of each scan
	s.exemptions.ClearCache()

	namespaces, err := s.getNamespacesToScan(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Info("scanning namespaces", "count", len(namespaces), "namespaces", namespaces)

	var findings []types.Finding

	// Run resource checks
	if s.config.ChecksResourcesEnabled {
		resourceFindings, err := s.checkResources(ctx, namespaces)
		if err != nil {
			s.logger.Error("resource checks failed", "error", err)
		} else {
			findings = append(findings, resourceFindings...)
		}
	}

	// Run probe checks
	if s.config.ChecksProbesEnabled {
		probeFindings, err := s.checkProbes(ctx, namespaces)
		if err != nil {
			s.logger.Error("probe checks failed", "error", err)
		} else {
			findings = append(findings, probeFindings...)
		}
	}

	// Run replica checks
	if s.config.ChecksReplicasEnabled {
		replicaFindings, err := s.checkReplicas(ctx, namespaces)
		if err != nil {
			s.logger.Error("replica checks failed", "error", err)
		} else {
			findings = append(findings, replicaFindings...)
		}
	}

	// Run image checks
	if s.config.ChecksImagesEnabled {
		imageFindings, err := s.checkImages(ctx, namespaces)
		if err != nil {
			s.logger.Error("image checks failed", "error", err)
		} else {
			findings = append(findings, imageFindings...)
		}
	}

	// Run security context checks
	if s.config.ChecksSecurityContextEnabled {
		securityFindings, err := s.checkSecurity(ctx, namespaces)
		if err != nil {
			s.logger.Error("security checks failed", "error", err)
		} else {
			findings = append(findings, securityFindings...)
		}
	}

	// Run deprecated API checks
	if s.config.ChecksDeprecatedAPIEnabled {
		deprecatedFindings, err := s.checkDeprecatedAPIs(ctx, namespaces)
		if err != nil {
			s.logger.Error("deprecated API checks failed", "error", err)
		} else {
			findings = append(findings, deprecatedFindings...)
		}
	}

	// Run namespace checks
	if s.config.ChecksNamespaceEnabled {
		namespaceFindings, err := s.checkNamespaces(ctx, namespaces)
		if err != nil {
			s.logger.Error("namespace checks failed", "error", err)
		} else {
			findings = append(findings, namespaceFindings...)
		}
	}

	s.logger.Info("scan completed", "findings", len(findings))
	return findings, nil
}

func (s *Scanner) getNamespacesToScan(ctx context.Context) ([]string, error) {
	// If specific namespaces are configured, use those
	if len(s.config.ScanNamespaces) > 0 {
		return s.config.ScanNamespaces, nil
	}

	// Otherwise, get all namespaces and filter out excluded ones
	nsList, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	excludeMap := make(map[string]bool)
	for _, ns := range s.config.ExcludeNamespaces {
		excludeMap[ns] = true
	}

	var namespaces []string
	for _, ns := range nsList.Items {
		if !excludeMap[ns.Name] {
			namespaces = append(namespaces, ns.Name)
		}
	}

	return namespaces, nil
}

func (s *Scanner) HealthCheck(ctx context.Context) error {
	_, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	return err
}
