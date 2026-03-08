package presenter

import (
	"context"
	"errors"
	"fmt"

	"github.com/lugassawan/panen/backend/usecase"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ExportImportHandler handles data export and import requests.
type ExportImportHandler struct {
	ctx       context.Context
	exportSvc *usecase.ExportService
	importSvc *usecase.ImportService
}

// Bind injects runtime dependencies into the handler.
func (h *ExportImportHandler) Bind(
	ctx context.Context,
	exportSvc *usecase.ExportService,
	importSvc *usecase.ImportService,
) {
	h.ctx = ctx
	h.exportSvc = exportSvc
	h.importSvc = importSvc
}

// ExportData prompts the user for a save location and exports the database.
func (h *ExportImportHandler) ExportData() (string, error) {
	path, err := wailsRuntime.SaveFileDialog(h.ctx, wailsRuntime.SaveDialogOptions{
		DefaultFilename: usecase.DefaultExportFilename(),
		Title:           "Export Data",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Zip Archives", Pattern: "*.zip"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("export data: %w", err)
	}
	if path == "" {
		return "", nil // user cancelled
	}

	if _, err := h.exportSvc.Export(path); err != nil {
		return "", fmt.Errorf("export data: %w", err)
	}
	return path, nil
}

// ImportPreview prompts the user to select a file and returns a preview.
func (h *ExportImportHandler) ImportPreview() (*ImportPreviewResponse, error) {
	path, err := wailsRuntime.OpenFileDialog(h.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Import Data",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Zip Archives", Pattern: "*.zip"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("import preview: %w", err)
	}
	if path == "" {
		return &ImportPreviewResponse{}, nil // user cancelled
	}

	result, err := h.importSvc.Preview(path)
	if err != nil {
		return nil, fmt.Errorf("import preview: %w", err)
	}

	return &ImportPreviewResponse{
		FilePath:   path,
		AppVersion: result.AppVersion,
		ExportedAt: result.ExportedAt,
		Checksum:   result.Checksum,
		DbSize:     result.DbSize,
	}, nil
}

// ImportData executes the import from the given archive file path.
func (h *ExportImportHandler) ImportData(filePath string) error {
	if filePath == "" {
		return errors.New("import data: file path is required")
	}
	if err := h.importSvc.Import(filePath); err != nil {
		return fmt.Errorf("import data: %w", err)
	}
	return nil
}
