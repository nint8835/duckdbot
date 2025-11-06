package embedding

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/pipelines"
)

var Pipeline *pipelines.FeatureExtractionPipeline

func Initialize() error {
	slog.Debug("Initializing embedding support")

	session, err := hugot.NewGoSession()
	if err != nil {
		return fmt.Errorf("failed to create Hugot session: %w", err)
	}

	modelsDir := "./models/"
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	downloadOptions := hugot.NewDownloadOptions()
	downloadOptions.OnnxFilePath = "onnx/model.onnx"

	slog.Debug("Downloading embedding model")
	modelPath, err := hugot.DownloadModel(
		"sentence-transformers/all-MiniLM-L6-v2",
		modelsDir,
		downloadOptions,
	)
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}

	config := hugot.FeatureExtractionConfig{
		ModelPath: modelPath,
		Name:      "embeddingPipeline",
	}

	slog.Debug("Creating embedding pipeline", "modelPath", modelPath)
	embeddingPipeline, err := hugot.NewPipeline(session, config)
	if err != nil {
		return fmt.Errorf("failed to create embedding pipeline: %w", err)
	}

	slog.Debug("Pipeline created successfully")

	Pipeline = embeddingPipeline
	return nil
}
