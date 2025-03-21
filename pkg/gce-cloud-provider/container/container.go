package gkecloudprovider

import (
	"context"
	"fmt"
	"strings"

	container "cloud.google.com/go/container/apiv1"
	containerpb "cloud.google.com/go/container/apiv1/containerpb"
	"google.golang.org/api/option"
	"sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/common"
	"sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/gce-cloud-provider/auth"
	"sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/gce-cloud-provider/metadata"
)

type GKEClusterProvider struct {
	service     *container.ClusterManagerClient
	projectName string
	clusterName string
	location    string
}

func NewGKEClusterProvider(ctx context.Context, conf *auth.AuthConfig, mdService metadata.MetadataService) (*GKEClusterProvider, error) {
	service, err := container.NewClusterManagerClient(ctx, option.WithTokenSource(conf.Token))
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster manager client: %v", err)
	}

	return &GKEClusterProvider{
		service:     service,
		projectName: mdService.GetProject(),
		location:    mdService.GetClusterLocation(),
		clusterName: mdService.GetClusterName(),
	}, nil
}

func (p *GKEClusterProvider) ListNodePools(ctx context.Context) ([]*containerpb.NodePool, error) {
	if p.projectName == "" || p.location == "" || p.clusterName == "" {
		return nil, fmt.Errorf("projectName, location, and clusterName must be set")
	}

	req := &containerpb.ListNodePoolsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/clusters/%s",
			p.projectName, p.location, p.clusterName),
	}

	resp, err := p.service.ListNodePools(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list node pools: %v", err)
	}

	return resp.NodePools, nil
}

// AllNodePoolsLabeled returns true if all node pools have a label that
// indicates they support the disk type.
func AllNodePoolsLabeled(ctx context.Context, nodepools []*containerpb.NodePool) (bool, error) {
	for _, np := range nodepools {
		conf := np.GetConfig()
		if conf == nil {
			return false, fmt.Errorf("node pool %s has no config", np.GetName())
		}

		labels := conf.GetLabels()
		if labels == nil {
			return false, fmt.Errorf("node pool %s has no labels", np.GetName())
		}

		if !hasMatchingLabel(labels) {
			return false, nil
		}
	}

	return true, nil
}

func hasMatchingLabel(labels map[string]string) bool {
	for k, _ := range labels {
		if matchesDiskSupportLabel(k) {
			return true
		}
	}

	return false
}

func matchesDiskSupportLabel(key string) bool {
	if !common.IsGKETopologyLabel(key) {
		return false
	}

	diskType := strings.TrimPrefix(key, common.TopologyKeyPrefix+"/")
	return strings.HasPrefix(diskType, "pd-") || strings.HasPrefix(diskType, "hyperdisk-")
}
