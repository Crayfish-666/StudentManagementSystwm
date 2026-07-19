package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// ApplicationRepository 入团申请数据访问层。
type ApplicationRepository struct {
	db *gorm.DB
}

// NewApplicationRepository 创建入团申请仓储。
func NewApplicationRepository(db *gorm.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

// List 分页查询入团申请列表，支持按状态/学生/院系/专业筛选。
func (r *ApplicationRepository) List(status string, studentID, collegeID int64, majorIDs []int64, page, pageSize int) ([]models.TyApplication, int64, error) {
	query := r.db.Where("is_deleted = 0")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if studentID > 0 {
		query = query.Where("student_id = ?", studentID)
	}
	if collegeID > 0 {
		query = query.Where("branch_id IN (SELECT id FROM ty_branch WHERE college_id = ?)", collegeID)
	}
	if len(majorIDs) > 0 {
		query = query.Where("student_id IN (SELECT id FROM idx_student WHERE major_id IN ? AND is_deleted = 0)", majorIDs)
	}

	var total int64
	if err := query.Model(&models.TyApplication{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var apps []models.TyApplication
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&apps).Error; err != nil {
		return nil, 0, err
	}

	return apps, total, nil
}

// GetByID 按 ID 查询单个入团申请。
func (r *ApplicationRepository) GetByID(id int64) (*models.TyApplication, error) {
	var app models.TyApplication
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// HasPending 检查学生是否存在 S1/S2 状态的申请（同一学生同一时间只允许 1 份）。
func (r *ApplicationRepository) HasPending(studentID int64) (bool, error) {
	var count int64
	if err := r.db.Model(&models.TyApplication{}).
		Where("student_id = ? AND status IN ('S1','S2') AND is_deleted = 0", studentID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasActiveApplication 全周期 1 单限制：每名学生终身只能有一条入团申请。
// 同一学生在任意时刻最多存在 1 条未删除的申请单（含 S0 草稿、S1/S2 审批中、S3 已通过）。
// S4 已驳回视为可重新提交，不限制。
func (r *ApplicationRepository) HasActiveApplication(studentID int64, excludeID int64) (bool, *models.TyApplication, error) {
	var app models.TyApplication
	err := r.db.Where("student_id = ? AND is_deleted = 0 AND status <> 'S4'", studentID).
		Where("id <> ?", excludeID).
		Order("id DESC").
		First(&app).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &app, nil
}

// Create 创建入团申请。
func (r *ApplicationRepository) Create(app *models.TyApplication) error {
	return r.db.Create(app).Error
}

// UpdateStudentPoliticalStatus 更新学生政治面貌（终审通过 S3 时自动调用）。
func (r *ApplicationRepository) UpdateStudentPoliticalStatus(studentID int64, status string) error {
	return r.db.Model(&models.IdxStudent{}).
		Where("id = ? AND is_deleted = 0", studentID).
		Update("political_status", status).Error
}

// Update 更新入团申请。
func (r *ApplicationRepository) Update(app *models.TyApplication) error {
	return r.db.Save(app).Error
}

// SoftDelete 软删除入团申请。
func (r *ApplicationRepository) SoftDelete(id int64) error {
	return r.db.Model(&models.TyApplication{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// GetStudentByID 查询学生信息（用于年龄校验）。
func (r *ApplicationRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// GetUserByID 查询用户信息（用于获取关联的 student_id）。
func (r *ApplicationRepository) GetUserByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetBranchByID 查询团支部信息。
func (r *ApplicationRepository) GetBranchByID(id int64) (*models.TyBranch, error) {
	var branch models.TyBranch
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

// ListBranchesByCollege 按院系查询团支部。
func (r *ApplicationRepository) ListBranchesByCollege(collegeID int64) ([]models.TyBranch, error) {
	var branches []models.TyBranch
	query := r.db.Where("is_deleted = 0")
	if collegeID > 0 {
		query = query.Where("college_id = ?", collegeID)
	}
	if err := query.Order("id ASC").Find(&branches).Error; err != nil {
		return nil, err
	}
	return branches, nil
}

// ListColleges 查询所有院系。
func (r *ApplicationRepository) ListColleges() ([]models.SysCollege, error) {
	var colleges []models.SysCollege
	if err := r.db.Where("is_deleted = 0").Order("id ASC").Find(&colleges).Error; err != nil {
		return nil, err
	}
	return colleges, nil
}

// CreateApprovalRecord 写入一条审批记录。
func (r *ApplicationRepository) CreateApprovalRecord(rec *models.TyApprovalRecord) error {
	return r.db.Create(rec).Error
}

// ListApprovalRecords 列出某申请单全部审批记录（时间正序）。
func (r *ApplicationRepository) ListApprovalRecords(applicationID int64) ([]models.TyApprovalRecord, error) {
	var records []models.TyApprovalRecord
	if err := r.db.Where("application_id = ? AND is_deleted = 0", applicationID).
		Order("occurred_at ASC, id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// HasApprovedStep 判断指定 step 是否已通过（用于校级前置校验院系是否完成）。
func (r *ApplicationRepository) HasApprovedStep(applicationID int64, step string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.TyApprovalRecord{}).
		Where("application_id = ? AND step = ? AND result = 'approve' AND is_deleted = 0", applicationID, step).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasApprovedStepByModule 判断指定模块+目标+步骤是否已通过（用于全流程审批前置校验）。
func (r *ApplicationRepository) HasApprovedStepByModule(applicationID int64, module string, targetID int64, step string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.TyApprovalRecord{}).
		Where("application_id = ? AND module = ? AND target_id = ? AND step = ? AND result = 'approve' AND is_deleted = 0",
			applicationID, module, targetID, step).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListApprovalRecordsByModule 按模块+目标ID查询审批记录。
func (r *ApplicationRepository) ListApprovalRecordsByModule(module string, targetID int64) ([]models.TyApprovalRecord, error) {
	var records []models.TyApprovalRecord
	if err := r.db.Where("module = ? AND target_id = ? AND is_deleted = 0", module, targetID).
		Order("occurred_at ASC, id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// ListAllApprovalRecordsByApplication 查询某申请单全流程所有审批记录（含所有模块）。
func (r *ApplicationRepository) ListAllApprovalRecordsByApplication(applicationID int64) ([]models.TyApprovalRecord, error) {
	var records []models.TyApprovalRecord
	if err := r.db.Where("application_id = ? AND is_deleted = 0", applicationID).
		Order("occurred_at ASC, id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// FindUserRoles 查询用户角色码列表。
func (r *ApplicationRepository) FindUserRoles(userID int64) ([]string, error) {
	type row struct {
		Code string
	}
	var rows []row
	if err := r.db.Table("sys_user_role AS ur").
		Select("r.code AS code").
		Joins("JOIN sys_role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND ur.is_deleted = 0 AND r.is_deleted = 0", userID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(rows))
	for _, x := range rows {
		codes = append(codes, x.Code)
	}
	return codes, nil
}

// FindUserScopeCollegeIDs 查询用户作用域内院系 ID 列表。
func (r *ApplicationRepository) FindUserScopeCollegeIDs(userID int64) ([]int64, error) {
	type row struct {
		CollegeID int64
	}
	var rows []row
	if err := r.db.Table("sys_user_role AS ur").
		Select("ur.scope_college_id AS college_id").
		Where("ur.user_id = ? AND ur.is_deleted = 0 AND ur.scope_college_id IS NOT NULL", userID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, x := range rows {
		ids = append(ids, x.CollegeID)
	}
	return ids, nil
}

// FindCounselorMajorIDs 查询辅导员负责的所有专业 ID 列表。
// 通过 idx_class.counselor_id = userID 关联到班级，再取班级所属专业 major_id。
func (r *ApplicationRepository) FindCounselorMajorIDs(userID int64) ([]int64, error) {
	type row struct {
		MajorID int64
	}
	var rows []row
	if err := r.db.Table("idx_class").
		Select("DISTINCT major_id AS major_id").
		Where("counselor_id = ? AND is_deleted = 0", userID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, x := range rows {
		ids = append(ids, x.MajorID)
	}
	return ids, nil
}
