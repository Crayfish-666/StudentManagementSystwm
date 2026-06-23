package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// FileRepository 文件元数据数据访问层。
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository 创建文件仓库。
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create 创建文件元数据记录。
func (r *FileRepository) Create(meta *models.FileMeta) error {
	return r.db.Create(meta).Error
}

// GetByID 根据 ID 查找文件元数据（未删除）。
func (r *FileRepository) GetByID(id int64) (*models.FileMeta, error) {
	var meta models.FileMeta
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&meta).Error
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

// GetByStorageKey 根据 storage_key 查找文件元数据（未删除）。
func (r *FileRepository) GetByStorageKey(key string) (*models.FileMeta, error) {
	var meta models.FileMeta
	err := r.db.Where("storage_key = ? AND is_deleted = 0", key).First(&meta).Error
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

// SoftDelete 软删除文件元数据。
func (r *FileRepository) SoftDelete(id int64) error {
	return r.db.Model(&models.FileMeta{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// ListByUploader 分页查询指定上传者的文件列表。
func (r *FileRepository) ListByUploader(uploaderID int64, page, pageSize int) ([]models.FileMeta, int64, error) {
	var list []models.FileMeta
	var total int64

	db := r.db.Where("uploader_id = ? AND is_deleted = 0", uploaderID)
	if err := db.Model(&models.FileMeta{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
