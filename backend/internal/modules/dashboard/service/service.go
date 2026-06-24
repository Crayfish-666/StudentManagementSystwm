// Package service 工作台统计服务——按用户角色返回差异化内容。
package service

import (
	"gorm.io/gorm"
	"student-system/internal/models"
)

// DashboardService 工作台统计服务。
type DashboardService struct {
	db *gorm.DB
}

// NewDashboardService 创建工作台服务。
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

// OverviewView 工作台概览。
type OverviewView struct {
	User       UserView    `json:"user"`
	RoleScope  string      `json:"role_scope"` // student / college / school
	Stats      StatsView   `json:"stats"`
	TodoItems  []TodoItem  `json:"todo_items"`
	QuickLinks []QuickLink `json:"quick_links"`
}

// UserView 用户信息。
type UserView struct {
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	Roles       []RoleItem `json:"roles"`
}

// RoleItem 角色项。
type RoleItem struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Scope string `json:"scope"`
}

// StatsView 统计数据（字段按角色动态填充，未使用的为 0）。
type StatsView struct {
	// 学生视角
	MyTyStatus            string `json:"my_ty_status"`            // 我的入团状态：none / applying / cultivating / member
	MyCmpScore            int64  `json:"my_cmp_score"`            // 我的综合分（最近学期）
	MyActivityCount       int64  `json:"my_activity_count"`       // 我参加的活动数
	UnreadNotiCount       int64  `json:"unread_noti_count"`       // 未读通知
	RecruitingPlanCount   int64  `json:"recruiting_plan_count"`   // 招新中计划数（st_recruit_plan.status='S3' AND is_finished=0）

	// 教师视角
	StudentCount       int64 `json:"student_count"`        // 管辖学生数
	TyPendingCount     int64 `json:"ty_pending_count"`     // 待审批入团申请
	IncidentOpenCount  int64 `json:"incident_open_count"`  // 未关闭事件
	QgPositionCount    int64 `json:"qg_position_count"`    // 在岗勤工岗位
	ActiveAssocCount   int64 `json:"active_assoc_count"`   // 活跃社团数
}

// TodoItem 待办事项。
type TodoItem struct {
	Title string `json:"title"`
	Count int64  `json:"count"`
	Path  string `json:"path"`
	Color string `json:"color"` // 可选：高亮色
}

// QuickLink 快捷入口。
type QuickLink struct {
	Title string `json:"title"`
	Path  string `json:"path"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// Overview 工作台概览——根据用户角色返回不同内容。
func (s *DashboardService) Overview(uid interface{}) (*OverviewView, error) {
	view := &OverviewView{}

	// 1. 查询用户基本信息
	var user models.SysUser
	if err := s.db.Where("id = ? AND is_deleted = 0", uid).First(&user).Error; err != nil {
		return nil, err
	}
	view.User = UserView{
		Username:    user.Username,
		DisplayName: user.DisplayName,
	}

	// 2. 查询角色列表，判断主角色域
	var roles []models.SysRole
	s.db.Raw(`SELECT r.* FROM sys_role r
		JOIN sys_user_role ur ON ur.role_id = r.id
		WHERE ur.user_id = ? AND r.is_deleted = 0 AND ur.is_deleted = 0`, user.ID).Scan(&roles)

	highestScope := "student" // 默认学生
	for _, r := range roles {
		view.User.Roles = append(view.User.Roles, RoleItem{Code: r.Code, Name: r.Name, Scope: r.Scope})
		// scope 优先级：school > college > student
		if r.Scope == "school" {
			highestScope = "school"
		} else if r.Scope == "college" && highestScope != "school" {
			highestScope = "college"
		}
	}
	view.RoleScope = highestScope

	// 3. 通用：未读通知
	s.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = 0", user.ID).Count(&view.Stats.UnreadNotiCount)

	// 4. 按角色域分支
	if highestScope == "student" {
		s.buildStudentView(view, &user)
	} else {
		s.buildStaffView(view, &user, roles)
	}

	return view, nil
}

// buildStudentView 构建学生工作台视图。
func (s *DashboardService) buildStudentView(view *OverviewView, user *models.SysUser) {
	// 我的入团申请状态
	if user.StudentID != nil {
		var app models.TyApplication
		if err := s.db.Where("student_id = ? AND is_deleted = 0", *user.StudentID).
			Order("id DESC").First(&app).Error; err == nil {
			view.Stats.MyTyStatus = app.Status
		}
	}

	// 我参加的活动数
	if user.StudentID != nil {
		s.db.Table("st_activity_checkins").
			Where("student_id = ? AND is_deleted = 0", *user.StudentID).
			Count(&view.Stats.MyActivityCount)
	}

	// 招新中计划数：status=S3（已通过/可投递）且 is_finished=0（招新中）
	s.db.Table("st_recruit_plan").
		Where("status = ? AND is_finished = 0 AND is_deleted = 0", "S3").
		Count(&view.Stats.RecruitingPlanCount)

	// 待办：我的入团申请进度、未读通知
	view.TodoItems = []TodoItem{}
	if view.Stats.MyTyStatus != "" && view.Stats.MyTyStatus != "S7_MEMBER" && view.Stats.MyTyStatus != "WITHDRAWN" {
		// 兼容短码（S1~S6）和长码（S1_SUBMITTED ~ S6_ADMITTED）两种格式
		statusLabel := map[string]string{
			"S1":             "入团申请已提交，等待审核",
			"S2":             "推优通过，进入培养考察",
			"S3":             "培养考察中",
			"S4":             "已列为发展对象",
			"S5":             "政审已完成",
			"S6":             "已被接收为预备团员",
			"S1_SUBMITTED":   "入团申请已提交，等待审核",
			"S2_RECOMMENDED": "推优通过，进入培养考察",
			"S3_CULTIVATING": "培养考察中",
			"S4_DEVELOPING":  "已列为发展对象",
			"S5_POLITICED":   "政审已完成",
			"S6_ADMITTED":    "已被接收为预备团员",
		}
		label := statusLabel[view.Stats.MyTyStatus]
		if label == "" {
			label = "入团发展进行中"
		}
		view.TodoItems = append(view.TodoItems, TodoItem{
			Title: label,
			Path:  "/mine/ty-application",
			Color: "#e6a23c",
		})
	}
	if view.Stats.UnreadNotiCount > 0 {
		view.TodoItems = append(view.TodoItems, TodoItem{
			Title: "未读通知",
			Count: view.Stats.UnreadNotiCount,
			Path:  "/notifications",
			Color: "#409eff",
		})
	}
	if len(view.TodoItems) == 0 {
		view.TodoItems = append(view.TodoItems, TodoItem{
			Title: "暂无待办",
			Path:  "",
		})
	}

	// 快捷入口：学生视角
	view.QuickLinks = []QuickLink{
		{Title: "团员发展", Path: "/ty/application", Icon: "Flag", Color: "#e6a23c"},
		{Title: "社团活动", Path: "/st/association", Icon: "Trophy", Color: "#67c23a"},
		{Title: "学生社区", Path: "/sq/building", Icon: "House", Color: "#409eff"},
		{Title: "勤工助学", Path: "/qg/difficulty", Icon: "Briefcase", Color: "#f56c6c"},
		{Title: "综合素质", Path: "/cmp/my-score", Icon: "TrendCharts", Color: "#9b59b6"},
		{Title: "我的申请", Path: "/mine/ty-application", Icon: "Document", Color: "#909399"},
	}
}

// buildStaffView 构建教师/管理员工作台视图。
//
// 规则：
//   - 当用户拥有 R-COL-COUN（院系辅导员）角色时，stats.student_count
//     统计其所带班级（idx_class.counselor_id = user.id）下的学生数；
//   - 其他教师/管理员维持全校学生总数。
func (s *DashboardService) buildStaffView(view *OverviewView, user *models.SysUser, roles []models.SysRole) {
	// 在校学生总数：辅导员展示"所带学生数"，其他角色展示全校学生数。
	isCounselor := false
	for _, r := range roles {
		if r.Code == "R-COL-COUN" {
			isCounselor = true
			break
		}
	}
	if isCounselor {
		s.db.Table("idx_student AS st").
			Joins("JOIN idx_class c ON c.id = st.class_id AND c.is_deleted = 0").
			Where("st.is_deleted = 0 AND c.counselor_id = ?", user.ID).
			Count(&view.Stats.StudentCount)
	} else {
		s.db.Model(&models.IdxStudent{}).Where("is_deleted = 0").Count(&view.Stats.StudentCount)
	}

	// 待审批入团申请
	s.db.Model(&models.TyApplication{}).Where("status IN ? AND is_deleted = 0",
		[]string{"S1_SUBMITTED", "S2_RECOMMENDED", "S3_CULTIVATING", "S4_DEVELOPING", "S5_POLITICED"}).
		Count(&view.Stats.TyPendingCount)

	// 未关闭社区事件
	s.db.Model(&models.SqIncident{}).Where("status NOT IN ? AND is_deleted = 0",
		[]string{"closed", "resolved"}).Count(&view.Stats.IncidentOpenCount)

	// 在岗勤工岗位
	s.db.Model(&models.QgPosition{}).Where("status = ? AND is_deleted = 0", "open").
		Count(&view.Stats.QgPositionCount)

	// 活跃社团
	s.db.Model(&models.StAssociation{}).Where("status IN ? AND is_deleted = 0",
		[]string{"registered", "trial"}).Count(&view.Stats.ActiveAssocCount)

	// 待办：审批任务 + 事件处理 + 通知
	view.TodoItems = []TodoItem{}
	if view.Stats.TyPendingCount > 0 {
		view.TodoItems = append(view.TodoItems, TodoItem{
			Title: "待审批入团申请",
			Count: view.Stats.TyPendingCount,
			Path:  "/ty/approval",
			Color: "#e6a23c",
		})
	}
	if view.Stats.IncidentOpenCount > 0 {
		view.TodoItems = append(view.TodoItems, TodoItem{
			Title: "待处理社区事件",
			Count: view.Stats.IncidentOpenCount,
			Path:  "/sq/incident",
			Color: "#f56c6c",
		})
	}
	if view.Stats.UnreadNotiCount > 0 {
		view.TodoItems = append(view.TodoItems, TodoItem{
			Title: "未读通知",
			Count: view.Stats.UnreadNotiCount,
			Path:  "/notifications",
			Color: "#409eff",
		})
	}
	if len(view.TodoItems) == 0 {
		view.TodoItems = append(view.TodoItems, TodoItem{Title: "暂无待办", Path: ""})
	}

	// 快捷入口：教师/管理员视角
	view.QuickLinks = []QuickLink{
		{Title: "入团审批", Path: "/ty/approval", Icon: "Flag", Color: "#e6a23c"},
		{Title: "社团管理", Path: "/st/association", Icon: "Trophy", Color: "#67c23a"},
		{Title: "社区管理", Path: "/sq/building", Icon: "House", Color: "#409eff"},
		{Title: "勤工管理", Path: "/qg/position", Icon: "Briefcase", Color: "#f56c6c"},
		{Title: "综合看板", Path: "/cmp/dashboard", Icon: "TrendCharts", Color: "#9b59b6"},
		{Title: "系统管理", Path: "/sys/user", Icon: "Setting", Color: "#606266"},
	}
}
