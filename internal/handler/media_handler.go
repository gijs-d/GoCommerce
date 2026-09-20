package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/config"
	"webshop/internal/storage"
)

type MediaHandler struct {
	storage storage.Storage
	cfg     *config.Config
}

func NewMediaHandler(storage storage.Storage, cfg *config.Config) *MediaHandler {
	return &MediaHandler{
		storage: storage,
		cfg:     cfg,
	}
}

type MediaFileInfo struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	SizeKB    int64     `json:"size_kb"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		api.Error(w, http.StatusBadRequest, "Bestand te groot (max 10MB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		api.Error(w, http.StatusBadRequest, "Geen bestand geselecteerd")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	url, err := h.storage.Upload(r.Context(), header.Filename, file, contentType)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Upload mislukt: "+err.Error())
		return
	}

	api.Success(w, http.StatusOK, "Bestand geüpload", map[string]string{
		"url":      url,
		"filename": header.Filename,
	})
}

// List haalt alle bestanden uit de uploads map op voor de Media Bibliotheek
func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	uploadDir := "./uploads"
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		api.JSON(w, http.StatusOK, []MediaFileInfo{})
		return
	}

	baseURL := strings.TrimSuffix(h.cfg.BaseURL, "/")
	var files []MediaFileInfo

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, MediaFileInfo{
			Name:      entry.Name(),
			URL:       fmt.Sprintf("%s/uploads/%s", baseURL, entry.Name()),
			SizeKB:    info.Size() / 1024,
			CreatedAt: info.ModTime(),
		})
	}

	api.JSON(w, http.StatusOK, files)
}

func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		api.Error(w, http.StatusBadRequest, "Bestandsnaam ontbreekt")
		return
	}

	targetPath := filepath.Join("./uploads", filepath.Base(filename))
	_ = os.Remove(targetPath)

	api.Success(w, http.StatusOK, "Bestand verwijderd", nil)
}