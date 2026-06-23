package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// InspectionRepository 巡查数据访问层。
type InspectionRepository struct {
	db *gorm.DB
}

// NewInspectionRepository 创建巡查仓储。
func NewInspectionRepository(db *gorm.DB) *InspectionRepository {
	return &InspectionRepository{db: db}
}

// List 分页查询巡查列表。
func (r *InspectionRepository) List(inspectionType string, buildingID int64, page, pageSize int) ([]models.SqInspection, int64, error) {
	query := r.db.Where("is_deleted = 0")

	if inspectionType != "" {
		query = query.Where("inspection_type = ?", inspectionType)
	}
	if buildingID > 0 {
		query = query.Where("building_id = ?", buildingID)
	}

	var total int64
	if err := query.Model(&models.SqInspection{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var inspections []models.SqInspection
	offset := (page - 1) * pageSize
	if err := query.Order("inspected_at DESC").Offset(offset).Limit(pageSize).Find(&inspections).Error; err != nil {
		return nil, 0, err
	}

	return inspections, total, nil
}

// GetByID 按 ID 查询巡查记录。
func (r *InspectionRepository) GetByID(id int64) (*models.SqInspection, error) {
	var insp models.SqInspection
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&insp).Error; err != nil {
		return nil, err
	}
	return &insp, nil
}

// Create 创建巡查记录。
func (r *InspectionRepository) Create(insp *models.SqInspection) error {
	return r.db.Create(insp).Error
}

// Update 更新巡查记录。
func (r *InspectionRepository) Update(insp *models.SqInspection) error {
	return r.db.Save(insp).Error
}

// SoftDelete 软删除巡查记录。
func (r *InspectionRepository) SoftDelete(id int64) error {
	return r.db.Model(&models.SqInspection{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// ---- 扣分项 ----

// ListDeductionsByInspection 查询巡查扣分项。
func (r *InspectionRepository) ListDeductionsByInspection(inspectionID int64) ([]models.SqInspectionDeduction, error) {
	var deductions []models.SqInspectionDeduction
	if err := r.db.Where("inspection_id = ?", inspectionID).Order("id ASC").Find(&deductions).Error; err != nil {
		return nil, err
	}
	return deductions, nil
}

// CreateDeduction 创建扣分项。
func (r *InspectionRepository) CreateDeduction(d *models.SqInspectionDeduction) error {
	return r.db.Create(d).Error
}

// DeleteDeductionsByInspection 删除巡查的所有扣分项（用于更新时先删后增）。
func (r *InspectionRepository) DeleteDeductionsByInspection(inspectionID int64) error {
	return r.db.Where("inspection_id = ?", inspectionID).Delete(&models.SqInspectionDeduction{}).Error
}

// ---- 辅助 ----

// GetBuildingByID 查询楼栋信息。
func (r *InspectionRepository) GetBuildingByID(id int64) (*models.IdxDormBuilding, error) {
	var b models.IdxDormBuilding
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

// GetFloorByID 查询楼层信息。
func (r *InspectionRepository) GetFloorByID(id int64) (*models.IdxDormFloor, error) {
	var f models.IdxDormFloor
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// GetRoomByID 查询寝室信息。
func (r *InspectionRepository) GetRoomByID(id int64) (*models.IdxDormRoom, error) {
	var room models.IdxDormRoom
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

// GetUserByID 查询用户信息。
func (r *InspectionRepository) GetUserByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
