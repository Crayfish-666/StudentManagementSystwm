package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"student-system/internal/models"
	"student-system/internal/modules/file/repository"
)

// FileService 文件业务服务层。
type FileService struct {
	repo    *repository.FileRepository
	storage Storage
	db      *gorm.DB
}

// NewFileService 创建文件服务。
func NewFileService(repo *repository.FileRepository, storage Storage, db *gorm.DB) *FileService {
	return &FileService{repo: repo, storage: storage, db: db}
}

// Upload 上传文件。
// 生成 storage_key 为 yyyy/mm/uuid.ext，计算 SHA256，写入 file_meta。
func (s *FileService) Upload(ctx context.Context, uploaderID int64, module, bizType, originalName string, file io.Reader, size int64) (*models.FileMeta, error) {
	// 生成 storage_key: yyyy/mm/uuid.ext
	now := time.Now()
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".bin"
	}
	storageKey := fmt.Sprintf("%s/%s/%s%s",
		now.Format("2006"),
		now.Format("01"),
		uuid.New().String(),
		ext,
	)

	// 读取全部内容，同时计算 SHA256
	var buf bytes.Buffer
	hash := sha256.New()
	// MultiWriter 同时写入 buffer 和 hash
	mw := io.MultiWriter(&buf, hash)
	if _, err := io.Copy(mw, file); err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}

	sha256Hex := fmt.Sprintf("%x", hash.Sum(nil))

	// 保存到存储
	if err := s.storage.Save(ctx, storageKey, &buf); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 推断 MIME 类型
	mimeType := mimeTypeFromExt(ext)

	meta := &models.FileMeta{
		Module:       module,
		BizType:      bizType,
		OriginalName: originalName,
		StorageKey:   storageKey,
		MimeType:     mimeType,
		SizeBytes:    size,
		SHA256:       sha256Hex,
		UploaderID:   uploaderID,
		Visibility:   "private",
		IsDeleted:    0,
	}

	if err := s.repo.Create(meta); err != nil {
		return nil, fmt.Errorf("写入文件元数据失败: %w", err)
	}

	return meta, nil
}

// Download 鉴权下载文件。
// private 只能创建者或 admin 下载，org 同组织可下载，public 所有人。
func (s *FileService) Download(ctx context.Context, id, userID int64, isAdmin bool) (io.ReadCloser, *models.FileMeta, error) {
	meta, err := s.repo.GetByID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("文件不存在")
	}

	// 鉴权
	switch meta.Visibility {
	case "private":
		if meta.UploaderID != userID && !isAdmin {
			return nil, nil, fmt.Errorf("无权访问该文件")
		}
	case "org":
		// 同组织判断：admin 或创建者可访问
		// TODO: 完整的 org 鉴权需要查询用户组织关系，当前简化为创建者或 admin
		if meta.UploaderID != userID && !isAdmin {
			return nil, nil, fmt.Errorf("无权访问该文件")
		}
	case "public":
		// 所有人可下载
	default:
		return nil, nil, fmt.Errorf("未知的可见性设置: %s", meta.Visibility)
	}

	reader, err := s.storage.Read(ctx, meta.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return reader, meta, nil
}

// Delete 软删除文件（仅创建者或 admin）。
func (s *FileService) Delete(ctx context.Context, id, userID int64, isAdmin bool) error {
	meta, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("文件不存在")
	}

	if meta.UploaderID != userID && !isAdmin {
		return fmt.Errorf("无权删除该文件")
	}

	return s.repo.SoftDelete(id)
}

// GetByID 获取文件元数据。
func (s *FileService) GetByID(ctx context.Context, id int64) (*models.FileMeta, error) {
	return s.repo.GetByID(id)
}

// GetByStorageKey 根据 storage_key 获取文件元数据。
func (s *FileService) GetByStorageKey(ctx context.Context, key string) (*models.FileMeta, error) {
	return s.repo.GetByStorageKey(key)
}

// mimeTypeFromExt 根据扩展名推断 MIME 类型。
func mimeTypeFromExt(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".txt":
		return "text/plain"
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".xml":
		return "text/xml"
	case ".html", ".htm":
		return "text/html"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
