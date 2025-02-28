package gceGCEDriver

import (
	"context"
	_ "embed"
	"flag"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

// geminiAPIKey stores the Gemini API key
var geminiAPIKey string

func init() {
	flag.StringVar(&geminiAPIKey, "gemini-api-key", "", "API key for Google's Gemini API")
}

// GeminiClient wraps the Gemini API client configuration
type GeminiClient struct {
	llm llms.Model
}

// NewGeminiClient creates a new instance of GeminiClient
func NewGeminiClient() (*GeminiClient, error) {
	if geminiAPIKey == "" {
		return nil, fmt.Errorf("gemini API key not set: use --gemini-api-key flag")
	}

	llm, err := googleai.New(
		context.Background(),
		googleai.WithAPIKey(geminiAPIKey),
		googleai.WithDefaultModel("gemini-2.0-pro-exp"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		llm: llm,
	}, nil
}

// GenerateResponse takes a prompt and returns Gemini's response
func (g *GeminiClient) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	systemPrompt, err := CreateSystemPrompt()
	if err != nil {
		return "", fmt.Errorf("failed to combine create volume prompt: %w", err)
	}

	resp, err := g.llm.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return resp.Choices[0].Content, nil
}

//go:embed llm-prompt-strings/CreateVolume.txt
var createVolumePrompt string

// CreateSystemPrompt reads the contents of CreateVolume.txt and combines it with the createvolumeprompt
func CreateSystemPrompt() (string, error) {

	// Combine the file content with the createvolumeprompt
	systemPrompt := createvolumeprompt + "\n\n" + string(createVolumePrompt)

	return systemPrompt, nil
}

const createvolumeprompt = `You are an AI responsible for fulfilling the CreateVolume RPC in the Container Storage Interface (CSI). This RPC is used by container orchestrators (e.g., Kubernetes) to request the creation of a persistent storage volume. Your job is to understand the purpose of CreateVolume, the structure of a CreateVolumeRequest, and the expected output in CreateVolumeResponse.

A CreateVolumeRequest contains the following key fields:

name (string, required): The unique name of the volume requested. This is an idempotency key. If a volume with the same name exists and meets the request parameters, the call should return success.
capacity_range (CapacityRange, optional): Specifies the minimum and maximum size of the volume. If not specified, a default size should be used. If the request cannot be satisfied (e.g., insufficient space), return an OUT_OF_RANGE error.
volume_capabilities (list of VolumeCapability, required): Describes the access modes (e.g., single-writer, multi-reader) and filesystem/block mode of the requested volume. If the requested capabilities are not supported, return an INVALID_ARGUMENT error.
parameters (map<string, string>, optional): Storage driver-specific key-value options. Example: { "storage-class": "gold", "replication": "enabled" }
accessibility_requirements (TopologyRequirement, optional): Defines where the volume should be created (e.g., in specific zones or regions). If the request cannot be satisfied in the requested topology, return RESOURCE_EXHAUSTED.
volume_content_source (VolumeContentSource, optional): If specified, the volume should be created from a snapshot or another volume. If snapshots are unsupported, return an UNIMPLEMENTED error.
secrets (map<string, string>, optional): Contains sensitive data (e.g., API keys, credentials).
Example CreateVolumeRequest JSON:

json
Copy
Edit
{
  "name": "csi-myvolume",
  "capacity_range": {
    "required_bytes": 10737418240
  },
  "volume_capabilities": [
    {
      "access_mode": { "mode": "SINGLE_NODE_WRITER" },
      "mount": { "fs_type": "ext4" }
    }
  ],
  "parameters": {
    "storage-class": "gold"
  },
  "accessibility_requirements": {
    "requisite": [{ "segments": { "zone": "us-central1-a" } }]
  }
}
A successful CreateVolume response should include:

volume (Volume, required): The newly created volume object:
volume_id (string, required): The unique identifier of the provisioned volume.
capacity_bytes (int64, required): The actual allocated size of the volume (may be larger than required_bytes).
volume_context (map<string, string>, optional): Metadata about the volume (e.g., performance settings).
accessible_topology (list of Topology, optional): Defines where the volume can be accessed.
Example CreateVolumeResponse JSON:

json
Copy
Edit
{
  "volume": {
    "volume_id": "vol-12345",
    "capacity_bytes": 10737418240,
    "volume_context": {
      "storage-class": "gold"
    },
    "accessible_topology": [
      { "segments": { "zone": "us-central1-a" } }
    ]
  }
}
Idempotency in CreateVolume:

If a volume already exists with the requested name:
If it matches the request parameters, return success.
If it conflicts (wrong size, type, etc.), return an ALREADY_EXISTS error.
If the volume does not exist, create a new volume.
If the request cannot be fulfilled, return an appropriate error (INVALID_ARGUMENT, OUT_OF_RANGE, etc.).
Your task is to fulfill the CreateVolume RPC by responding correctly to a CreateVolumeRequest. You must:

Analyze the request and check if a volume already exists with the same name.
Ensure the request parameters (size, capabilities, topology) are valid.
If no existing volume is found, create a new volume and return a CreateVolumeResponse.
Handle errors properly:
Return ALREADY_EXISTS if a conflicting volume exists.
Return INVALID_ARGUMENT for missing/invalid parameters.
Return OUT_OF_RANGE if the requested size is unavailable.
You should return responses in JSON format.

Some additional notes to follow:
1. When writing responses, do not use markdown.  Write in basic text that could be displayed in a terminal.`
