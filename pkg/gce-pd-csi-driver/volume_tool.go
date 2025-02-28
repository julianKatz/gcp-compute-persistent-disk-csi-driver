package gceGCEDriver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/tools"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

var _ tools.Tool = &VolumeChecker{}

// VolumeChecker implements a tool for checking volume existence
type VolumeChecker struct{}

// Name returns the name of the tool
func (t *VolumeChecker) Name() string {
	return "check_volume_exists"
}

// Description returns what the tool does
func (t *VolumeChecker) Description() string {
	return "Checks if a volume exists with the given name and returns its details if found"
}

// VolumeDetails represents the structure of a volume's information
type VolumeDetails struct {
	VolumeID           string              `json:"volume_id"`
	CapacityBytes      int64               `json:"capacity_bytes"`
	VolumeContext      map[string]string   `json:"volume_context,omitempty"`
	AccessibleTopology []map[string]string `json:"accessible_topology,omitempty"`
}

// Call executes the volume check
func (t *VolumeChecker) Call(ctx context.Context, input string) (string, error) {
	// For now, we'll simulate checking for a volume
	// In a real implementation, this would query your actual storage backend

	// Parse the input as volume name
	var volumeName string
	if err := json.Unmarshal([]byte(input), &volumeName); err != nil {
		return "", fmt.Errorf("invalid input format: %w", err)
	}

	// Simulate checking for volume existence
	// In reality, you would query your storage system here
	// Use GCE API to check if the volume exists
	computeService, err := compute.NewService(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create compute service: %w", err)
	}

	projectID := "your-gcp-project-id" // Replace with your GCP project ID
	zone := "us-central1-b"            // Replace with the appropriate zone

	// Get the disk information from GCE
	disk, err := computeService.Disks.Get(projectID, zone, volumeName).Context(ctx).Do()
	if err != nil {
		if googleapi.IsNotModified(err) {
			// Volume not found
			return "", nil
		}
		return "", fmt.Errorf("failed to get disk details: %w", err)
	}

	// Construct the volume details from the disk information
	details := VolumeDetails{
		VolumeID:      disk.Name,
		CapacityBytes: disk.SizeGb * 1024 * 1024 * 1024, // Convert GB to bytes
		VolumeContext: map[string]string{
			"type": disk.Type,
		},
		AccessibleTopology: []map[string]string{
			{"topology.gke.io/zone": zone},
		},
	}

	response, err := json.Marshal(details)
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume details: %w", err)
	}
	return string(response), nil
}
