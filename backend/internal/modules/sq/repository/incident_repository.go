package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// IncidentRepository 异常事件数据访问层。
type IncidentRepository struct {
	db *gorm.DB
}

// NewIncidentRepository 创建事件仓储。
func NewIncidentRepository(db *gorm.DB) *IncidentRepository {
	return &IncidentRepository{db: db}
}

// List 分页查询事件列表。
func (r *IncidentRepository) List(incidentLevel string, status string, buildingID int64, page, pageSize int) ([]models.SqIncident, int64, error) {
	query := r.db.Where("is_deleted = 0")

	if incidentLevel != "" {
		query = query.Where("incident_level = ?", incidentLevel)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if buildingID > 0 {
		query = query.Where("building_id = ?", buildingID)
	}

	var total int64
	if err := query.Model(&models.SqIncident{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var incidents []models.SqIncident
	offset := (page - 1) * pageSize
	if err := query.Order("occurred_at DESC").Offset(offset).Limit(pageSize).Find(&incidents).Error; err != nil {
		return nil, 0, err
	}

	return incidents, total, nil
}

// GetByID 按 ID 查询事件。
func (r *IncidentRepository) GetByID(id int64) (*models.SqIncident, error) {
	var inc models.SqIncident
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&inc).Error; err != nil {
		return nil, err
	}
	return &inc, nil
}

// Create 创建事件。
func (r *IncidentRepository) Create(inc *models.SqIncident) error {
	return r.db.Create(inc).Error
}

// Update 更新事件。
func (r *IncidentRepository) Update(inc *models.SqIncident) error {
	return r.db.Save(inc).Error
}

// SoftDelete 软删除事件。
func (r *IncidentRepository) SoftDelete(id int64) error {
	return r.db.Model(&models.SqIncident{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// ---- 附件 ----

// ListAttachmentsByIncident 查询事件附件。
func (r *IncidentRepository) ListAttachmentsByIncident(incidentID int64) ([]models.SqIncidentAttach, error) {
	var attaches []models.SqIncidentAttach
	if err := r.db.Where("incident_id = ?", incidentID).Order("id ASC").Find(&attaches).Error; err != nil {
		return nil, err
	}
	return attaches, nil
}

// CreateAttachment 创建事件附件。
func (r *IncidentRepository) CreateAttachment(a *models.SqIncidentAttach) error {
	return r.db.Create(a).Error
}

// ---- 处置记录 ----

// ListActionsByIncident 查询事件处置记录。
func (r *IncidentRepository) ListActionsByIncident(incidentID int64) ([]models.SqIncidentAction, error) {
	var actions []models.SqIncidentAction
	if err := r.db.Where("incident_id = ?", incidentID).Order("action_at ASC").Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// CreateAction 创建处置记录。
func (r *IncidentRepository) CreateAction(a *models.SqIncidentAction) error {
	return r.db.Create(a).Error
}

// ---- 辅助 ----

// GetBuildingByID 查询楼栋信息。
func (r *IncidentRepository) GetBuildingByID(id int64) (*models.IdxDormBuilding, error) {
	var b models.IdxDormBuilding
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

// GetUserByID 查询用户信息。
func (r *IncidentRepository) GetUserByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetStudentByID 查询学生信息。
func (r *IncidentRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}
