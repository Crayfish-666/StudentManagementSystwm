package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// BuildingRepository 楼栋/楼层/寝室数据访问层。
type BuildingRepository struct {
	db *gorm.DB
}

// NewBuildingRepository 创建楼栋仓储。
func NewBuildingRepository(db *gorm.DB) *BuildingRepository {
	return &BuildingRepository{db: db}
}

// ---- 楼栋 ----

// ListBuildings 查询楼栋列表。
func (r *BuildingRepository) ListBuildings() ([]models.IdxDormBuilding, error) {
	var buildings []models.IdxDormBuilding
	if err := r.db.Where("is_deleted = 0").Order("id ASC").Find(&buildings).Error; err != nil {
		return nil, err
	}
	return buildings, nil
}

// GetBuilding 按 ID 查询楼栋。
func (r *BuildingRepository) GetBuilding(id int64) (*models.IdxDormBuilding, error) {
	var b models.IdxDormBuilding
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

// CreateBuilding 创建楼栋。
func (r *BuildingRepository) CreateBuilding(b *models.IdxDormBuilding) error {
	return r.db.Create(b).Error
}

// UpdateBuilding 更新楼栋。
func (r *BuildingRepository) UpdateBuilding(b *models.IdxDormBuilding) error {
	return r.db.Save(b).Error
}

// SoftDeleteBuilding 软删除楼栋。
func (r *BuildingRepository) SoftDeleteBuilding(id int64) error {
	return r.db.Model(&models.IdxDormBuilding{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// CountBuildingByCode 按 code 统计楼栋数。
func (r *BuildingRepository) CountBuildingByCode(code string, excludeID int64) (int64, error) {
	var count int64
	query := r.db.Model(&models.IdxDormBuilding{}).Where("code = ? AND is_deleted = 0", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ---- 楼层 ----

// ListFloorsByBuilding 查询楼栋下所有楼层。
func (r *BuildingRepository) ListFloorsByBuilding(buildingID int64) ([]models.IdxDormFloor, error) {
	var floors []models.IdxDormFloor
	if err := r.db.Where("building_id = ? AND is_deleted = 0", buildingID).Order("floor_no ASC").Find(&floors).Error; err != nil {
		return nil, err
	}
	return floors, nil
}

// GetFloor 按 ID 查询楼层。
func (r *BuildingRepository) GetFloor(id int64) (*models.IdxDormFloor, error) {
	var f models.IdxDormFloor
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// CreateFloor 创建楼层。
func (r *BuildingRepository) CreateFloor(f *models.IdxDormFloor) error {
	return r.db.Create(f).Error
}

// UpdateFloor 更新楼层。
func (r *BuildingRepository) UpdateFloor(f *models.IdxDormFloor) error {
	return r.db.Save(f).Error
}

// SoftDeleteFloor 软删除楼层。
func (r *BuildingRepository) SoftDeleteFloor(id int64) error {
	return r.db.Model(&models.IdxDormFloor{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// ---- 寝室 ----

// ListRoomsByBuilding 查询楼栋下所有寝室。
func (r *BuildingRepository) ListRoomsByBuilding(buildingID int64) ([]models.IdxDormRoom, error) {
	var rooms []models.IdxDormRoom
	if err := r.db.Where("building_id = ? AND is_deleted = 0", buildingID).Order("room_no ASC").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

// ListRoomsByFloor 查询楼层下所有寝室。
func (r *BuildingRepository) ListRoomsByFloor(floorID int64) ([]models.IdxDormRoom, error) {
	var rooms []models.IdxDormRoom
	if err := r.db.Where("floor_id = ? AND is_deleted = 0", floorID).Order("room_no ASC").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

// GetRoom 按 ID 查询寝室。
func (r *BuildingRepository) GetRoom(id int64) (*models.IdxDormRoom, error) {
	var room models.IdxDormRoom
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

// CreateRoom 创建寝室。
func (r *BuildingRepository) CreateRoom(room *models.IdxDormRoom) error {
	return r.db.Create(room).Error
}

// UpdateRoom 更新寝室。
func (r *BuildingRepository) UpdateRoom(room *models.IdxDormRoom) error {
	return r.db.Save(room).Error
}

// SoftDeleteRoom 软删除寝室。
func (r *BuildingRepository) SoftDeleteRoom(id int64) error {
	return r.db.Model(&models.IdxDormRoom{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// CountRoomByNo 按楼栋+房间号统计寝室数。
func (r *BuildingRepository) CountRoomByNo(buildingID int64, roomNo string, excludeID int64) (int64, error) {
	var count int64
	query := r.db.Model(&models.IdxDormRoom{}).Where("building_id = ? AND room_no = ? AND is_deleted = 0", buildingID, roomNo)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ---- 床位 ----

// ListBedsByRoom 查询寝室下所有床位。
func (r *BuildingRepository) ListBedsByRoom(roomID int64) ([]models.IdxDormBed, error) {
	var beds []models.IdxDormBed
	if err := r.db.Where("room_id = ? AND is_deleted = 0", roomID).Order("bed_no ASC").Find(&beds).Error; err != nil {
		return nil, err
	}
	return beds, nil
}

// CreateBed 创建床位。
func (r *BuildingRepository) CreateBed(bed *models.IdxDormBed) error {
	return r.db.Create(bed).Error
}

// UpdateBed 更新床位。
func (r *BuildingRepository) UpdateBed(bed *models.IdxDormBed) error {
	return r.db.Save(bed).Error
}

// ---- 辅助 ----

// GetStudentByID 查询学生信息。
func (r *BuildingRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// GetUserByID 查询用户信息。
func (r *BuildingRepository) GetUserByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ListRoomMembers 查询寝室入住成员（通过床位表）。
func (r *BuildingRepository) ListRoomMembers(roomID int64) ([]models.IdxDormBed, error) {
	var beds []models.IdxDormBed
	if err := r.db.Where("room_id = ? AND is_deleted = 0 AND occupant_student_id IS NOT NULL", roomID).
		Order("bed_no ASC").Find(&beds).Error; err != nil {
		return nil, err
	}
	return beds, nil
}
