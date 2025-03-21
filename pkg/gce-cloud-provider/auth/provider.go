package auth

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"gopkg.in/gcfg.v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/gce-cloud-provider/metadata"
)

var tokenScopes = []string{
	compute.CloudPlatformScope,
	compute.ComputeScope,
}

type AuthConfig struct {
	Token   oauth2.TokenSource
	Project string
	Zone    string
}

func ConfigFromFile(ctx context.Context, path string, mdService metadata.MetadataService) (*AuthConfig, error) {
	configFile, err := readConfigFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	klog.V(2).Infof("Using GCE provider config %+v", configFile)

	token, err := generateTokenSource(ctx, configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token source: %w", err)
	}

	project, zone := getProjectAndZone(configFile, mdService)

	return &AuthConfig{
		Token:   token,
		Project: project,
		Zone:    zone,
	}, nil
}

func generateTokenSource(ctx context.Context, configFile *configFile) (oauth2.TokenSource, error) {
	if configFile != nil && configFile.global.TokenURL != "" && configFile.global.TokenURL != "nil" {
		// configFile.Global.TokenURL is defined
		// Use AltTokenSource

		tokenSource := NewAltTokenSource(configFile.global.TokenURL, configFile.global.TokenBody)
		klog.V(2).Infof("Using AltTokenSource %#v", tokenSource)
		return tokenSource, nil
	}

	// Use DefaultTokenSource
	tokenSource, err := google.DefaultTokenSource(
		ctx,
		tokenScopes...)

	// DefaultTokenSource relies on GOOGLE_APPLICATION_CREDENTIALS env var being set.
	if gac, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		klog.V(2).Infof("GOOGLE_APPLICATION_CREDENTIALS env var set %v", gac)
	} else {
		klog.Warningf("GOOGLE_APPLICATION_CREDENTIALS env var not set")
	}
	klog.V(2).Infof("Using DefaultTokenSource %#v", tokenSource)

	return tokenSource, err
}

type configFile struct {
	global configGlobal `gcfg:"global"`
}

type configGlobal struct {
	TokenURL  string `gcfg:"token-url"`
	TokenBody string `gcfg:"token-body"`
	ProjectId string `gcfg:"project-id"`
	Zone      string `gcfg:"zone"`
}

func readConfigFile(configPath string) (*configFile, error) {
	if configPath == "" {
		return nil, nil
	}

	reader, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't open cloud provider configuration at %s: %w", configPath, err)
	}
	defer reader.Close()

	cfg := &configFile{}
	if err := gcfg.FatalOnly(gcfg.ReadInto(cfg, reader)); err != nil {
		return nil, fmt.Errorf("couldn't read cloud provider configuration at %s: %w", configPath, err)
	}
	return cfg, nil
}

func getProjectAndZone(config *configFile, mdService metadata.MetadataService) (string, string) {
	var zone string
	if config == nil || config.global.Zone == "" {
		zone = mdService.GetZone()
		klog.V(2).Infof("Using GCP zone from the Metadata server: %q", zone)
	} else {
		zone = config.global.Zone
		klog.V(2).Infof("Using GCP zone from the local GCE cloud provider config file: %q", zone)
	}

	var projectID string
	if config == nil || config.global.ProjectId == "" {
		// Project ID is not available from the local GCE cloud provider config file.
		// This could happen if the driver is not running in the master VM.
		// Defaulting to project ID from the Metadata server.
		projectID = mdService.GetProject()
		klog.V(2).Infof("Using GCP project ID from the Metadata server: %q", projectID)
	} else {
		projectID = config.global.ProjectId
		klog.V(2).Infof("Using GCP project ID from the local GCE cloud provider config file: %q", projectID)
	}

	return projectID, zone
}
