package gceGCEDriver

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Parse flags before running tests to handle the gemini-api-key flag
	flag.Parse()

	// If API key not provided via flag, try to get it from env var
	if geminiAPIKey == "" {
		geminiAPIKey = os.Getenv("GEMINI_API_KEY")
	}

	os.Exit(m.Run())
}

const dummyCreateVolumeRequest = `{
    "name": "pvc-aef6a7bf-f29f-48e5-b045-02c62458faf7",
    "capacity_range": {
      "required_bytes": 1073741824
    },
    "volume_capabilities": [
      {
        "mount": {
          "fs_type": "ext4"
        },
        "access_mode": {
          "mode": "SINGLE_NODE_WRITER"
        }
      }
    ],
    "parameters": {
      "csi.storage.k8s.io/pv/name": "pvc-aef6a7bf-f29f-48e5-b045-02c62458faf7",
      "csi.storage.k8s.io/pvc/name": "my-pvc",
      "csi.storage.k8s.io/pvc/namespace": "default",
      "type": "pd-balanced"
    },
    "accessibility_requirements": {
      "requisite": [
        {
          "segments": {
            "topology.gke.io/zone": "us-central1-b"
          }
        },
        {
          "segments": {
            "topology.gke.io/zone": "us-central1-c"
          }
        },
        {
          "segments": {
            "topology.gke.io/zone": "us-central1-f"
          }
        }
      ],
      "preferred": [
        {
          "segments": {
            "topology.gke.io/zone": "us-central1-f"
          }
        },
        {
          "segments": {
            "topology.gke.io/zone": "us-central1-b"
          }
        },
        {
          "segments": {
            "topology.gke.io/zone": "us-central1-c"
          }
        }
      ]
    }
  }`

func TestGenerateResponse(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test - GEMINI_API_KEY environment variable not set")
	}

	geminiAPIKey = apiKey
	client, err := NewGeminiClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	prompt := fmt.Sprintf("Tell me what abstract LLM-style tools you'd need to fulfill a createVolumeRequest like this one: \n %s \n I will implement those tools for you, but I'd like to understand what you need to do in your terms", dummyCreateVolumeRequest)
	response, err := client.GenerateResponse(context.Background(), prompt)
	if err != nil {
		t.Fatalf("Failed to generate response: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	fmt.Println(response)
}
