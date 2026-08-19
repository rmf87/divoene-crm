package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/internal/core/services"
)

// BackupHandler provides manager-only backup endpoints.
type BackupHandler struct {
	svc *services.BackupService
}

// NewBackupHandler creates a new BackupHandler.
func NewBackupHandler(svc *services.BackupService) *BackupHandler {
	return &BackupHandler{svc: svc}
}

// RegisterBackupRoutes registers manager-only backup routes.
func RegisterBackupRoutes(rg *gin.RouterGroup, h *BackupHandler) {
	rg.POST("", h.createBackup)
	rg.GET("", h.listBackups)
	rg.GET("/meta", h.backupMeta)
	rg.POST("/upload", h.uploadBackup)
	rg.POST("/download", h.downloadBackups)
	rg.GET("/:name", h.downloadBackup)
	rg.POST("/:name/restore", h.restoreBackup)
	rg.DELETE("/:name", h.deleteBackup)
}

func (h *BackupHandler) createBackup(c *gin.Context) {
	info, err := h.svc.Create(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, info)
}

func (h *BackupHandler) listBackups(c *gin.Context) {
	files, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, files)
}

func (h *BackupHandler) backupMeta(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Meta(c.Request.Context()))
}

func (h *BackupHandler) downloadBackup(c *gin.Context) {
	name := c.Param("name")
	path, err := h.svc.Path(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+name)
	c.File(path)
}

// downloadBackups streams a ZIP containing the requested snapshots.
// Body: {"names": ["divoene-...sqlite3", ...]}
func (h *BackupHandler) downloadBackups(c *gin.Context) {
	var req struct {
		Names []string `json:"names"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if len(req.Names) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "names is required"})
		return
	}

	var paths []string
	for _, name := range req.Names {
		path, err := h.svc.Path(c.Request.Context(), name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		paths = append(paths, path)
	}

	filename := fmt.Sprintf("divoene-backups-%s.zip", time.Now().Format("20060102-150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/zip")

	zw := zip.NewWriter(c.Writer)
	for _, path := range paths {
		in, err := os.Open(path)
		if err != nil {
			continue
		}
		w, err := zw.Create(filepath.Base(path))
		if err != nil {
			in.Close()
			continue
		}
		_, _ = io.Copy(w, in)
		in.Close()
	}
	zw.Close()
}

func (h *BackupHandler) deleteBackup(c *gin.Context) {
	name := c.Param("name")
	if err := h.svc.Delete(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": name})
}

// uploadBackup accepts one or more files under the "files" field. Each file may
// be a single SQLite database or a ZIP archive containing several. The legacy
// single "file" field is also accepted. All valid uploads are kept in history.
func (h *BackupHandler) uploadBackup(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multipart form required"})
		return
	}

	var files []*multipart.FileHeader
	if len(form.File["files"]) > 0 {
		files = form.File["files"]
	} else if single, err := c.FormFile("file"); err == nil {
		files = []*multipart.FileHeader{single}
	}
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "files field is required"})
		return
	}

	var saved []gin.H
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			continue
		}

		if strings.HasSuffix(strings.ToLower(fh.Filename), ".zip") {
			items, err := h.svc.SaveZipFromReader(c.Request.Context(), src, fh.Filename)
			src.Close()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			for _, item := range items {
				saved = append(saved, gin.H{
					"name":        item.Name,
					"size":        item.Size,
					"modified_at": item.ModTime,
				})
			}
			continue
		}

		item, err := h.svc.SaveFromReader(c.Request.Context(), src, fh.Filename)
		src.Close()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		saved = append(saved, gin.H{
			"name":        item.Name,
			"size":        item.Size,
			"modified_at": item.ModTime,
		})
	}

	c.JSON(http.StatusCreated, gin.H{"saved": saved})
}

func (h *BackupHandler) restoreBackup(c *gin.Context) {
	name := c.Param("name")
	if err := h.svc.RestoreLive(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"restored": name,
		"message":  "Database restored. Server restarting...",
	})
}
