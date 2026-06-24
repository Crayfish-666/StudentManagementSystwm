// 一站式测试数据灌入脚本：为 TY / ST / SQ / QG 四大业务模块批量补足测试数据。
// 用法：在 backend 目录执行 `go run ./cmd/seedall`。
//
// 设计原则（与项目铁律 docs/01-05 对齐）：
//   - 仅追加，不清空：每张主表先 COUNT，若已 ≥ 目标数则跳过该表；
//   - 严格遵循 docs/03 表结构与 CHECK 约束（字段名、类型、长度、枚举值）；
//   - 覆盖各业务的关键状态/等级/类型，便于前端做筛选/分页/统计演示；
//   - 业务编号全部走 idgen.NextBizNo，避免手写重复；
//   - 加密字段（身份证/手机号/银行卡末四位）走 cryptox.Encrypt。
//
// 各模块规模（中等档）：
//   - TY：ty_branch +3、ty_application +8（覆盖 S0-S4）、审批 +12、推优 +2、培养 +6、团课 +6、思想汇报 +4、
//                  发展对象 +2、政审 +4、发展大会 +2、预备期 +3、转正 +1、花名册 +3。
//   - ST：st_association +4、st_activity +8（覆盖 A-D）、活动审批 +24、签到 +30、总结 +3、经费 +2、评优 +1。
//   - SQ：sq_selfgov_position +5、sq_inspection +6、sq_incident +5（覆盖 L1-L4）、晚归 +6、违规 +3、考核 +2。
//   - QG：qg_difficulty_cert +5（覆盖 4 等级）、qg_position +5、qg_position_apply +4、qg_attendance +30、
//                  qg_monthly_assess +3、qg_payroll +3。
package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/pkg/cryptox"
)

// ctx 共享上下文，避免每个函数重复传 db。
type ctx struct {
	db *gorm.DB
	// 缓存已查到的 ID，避免反复查表
	colleges   []models.SysCollege
	majors     []models.SysMajor
	classes    []models.IdxClass
	students   []models.IdxStudent
	buildings  []models.IdxDormBuilding
	floors     []models.IdxDormFloor
	rooms      []models.IdxDormRoom
	beds       []models.IdxDormBed
	roles      map[string]models.SysRole // code -> role
	users      []models.SysUser
	assoc      []models.StAssociation
	branches   []models.TyBranch
	positions  []models.QgPosition
	applys     []models.QgPositionApply
	now        time.Time
}

// resolveDBPath 与 server 进程保持同一份 SQLite 文件。
// 优先解析 configs/config.yaml 中的 db.path；
// 失败则在常见候选中选「数据量最大」的一份。
func resolveDBPath() string {
	// 1) 读 configs/config.yaml 的 db.path（按当前 cwd 解析，与 server 一致）
	for _, cfgPath := range []string{
		"configs/config.yaml",
		"../configs/config.yaml",
		"../../configs/config.yaml",
	} {
		if data, err := os.ReadFile(cfgPath); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				// 匹配  path: ./data/studenthub.db
				if strings.HasPrefix(line, "path:") {
					raw := strings.TrimSpace(strings.TrimPrefix(line, "path:"))
					raw = strings.Trim(raw, `"`)
					if raw != "" {
						if filepath.IsAbs(raw) {
							return raw
						}
						return filepath.Clean(raw)
					}
				}
			}
		}
	}

	// 2) 兜底：选数据量最大的那份
	candidates := []string{
		"data/studenthub.db",
		"../data/studenthub.db",
		"data/data/studenthub.db",
		"../data/data/studenthub.db",
	}
	type info struct {
		path string
		size int64
	}
	var best info
	best.size = -1
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			if st.Size() > best.size {
				best = info{path: c, size: st.Size()}
			}
		}
	}
	if best.path != "" {
		abs, _ := filepath.Abs(best.path)
		return abs
	}
	log.Fatal("找不到 studenthub.db，请确认 configs/config.yaml 或 data/studenthub.db 是否存在")
	return ""
}

func main() {
	dbPath := resolveDBPath()
	fmt.Printf("=== seedall 使用数据库: %s\n\n", dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	c := &ctx{db: db, now: time.Now()}
	c.loadAll()
	c.syncBizSeq()

	fmt.Println("=== 1) TY 团员发展 ===")
	if err := c.seedTY(); err != nil {
		log.Fatalf("seedTY: %v", err)
	}

	fmt.Println("\n=== 2) ST 社团活动 ===")
	if err := c.seedST(); err != nil {
		log.Fatalf("seedST: %v", err)
	}

	fmt.Println("\n=== 3) SQ 学生社区 ===")
	if err := c.seedSQ(); err != nil {
		log.Fatalf("seedSQ: %v", err)
	}

	fmt.Println("\n=== 4) QG 勤工助学 ===")
	if err := c.seedQG(); err != nil {
		log.Fatalf("seedQG: %v", err)
	}

	fmt.Println("\n=== 完成 ===")
	c.printSummary()
}

// loadAll 一次性加载所有基础数据 ID 到 ctx 缓存。
func (c *ctx) loadAll() {
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.colleges).Error; err != nil {
		log.Fatalf("拉取院系失败: %v", err)
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.majors).Error; err != nil {
		log.Fatalf("拉取专业失败: %v", err)
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.classes).Error; err != nil {
		log.Fatalf("拉取班级失败: %v", err)
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.students).Error; err != nil {
		log.Fatalf("拉取学生失败: %v", err)
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.buildings).Error; err != nil {
		log.Fatalf("拉取楼栋失败: %v", err)
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.floors).Error; err != nil {
		log.Fatalf("拉取楼层失败: %v", err)
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.rooms).Error; err != nil {
		log.Fatalf("拉取寝室失败: %v", err)
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.beds).Error; err != nil {
		log.Fatalf("拉取床位失败: %v", err)
	}
	c.roles = map[string]models.SysRole{}
	var roles []models.SysRole
	if err := c.db.Where("is_deleted = 0").Find(&roles).Error; err != nil {
		log.Fatalf("拉取角色失败: %v", err)
	}
	for _, r := range roles {
		c.roles[r.Code] = r
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.users).Error; err != nil {
		log.Fatalf("拉取用户失败: %v", err)
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.assoc).Error; err != nil {
		log.Fatalf("拉取社团失败: %v", err)
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.branches).Error; err != nil {
		log.Fatalf("拉取团支部失败: %v", err)
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.positions).Error; err != nil {
		log.Fatalf("拉取岗位失败: %v", err)
	}
	if err := c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.applys).Error; err != nil {
		log.Fatalf("拉取岗位申请失败: %v", err)
	}

	fmt.Printf("基础数据：院系=%d 专业=%d 班级=%d 学生=%d 楼栋=%d 楼层=%d 寝室=%d 床位=%d 角色=%d 用户=%d 社团=%d 团支部=%d 岗位=%d 岗位申请=%d\n",
		len(c.colleges), len(c.majors), len(c.classes), len(c.students), len(c.buildings),
		len(c.floors), len(c.rooms), len(c.beds), len(c.roles), len(c.users), len(c.assoc), len(c.branches),
		len(c.positions), len(c.applys))
}

// pickStudent 随机取一个学生。
func (c *ctx) pickStudent(skip map[int64]struct{}) (models.IdxStudent, bool) {
	if len(c.students) == 0 {
		return models.IdxStudent{}, false
	}
	for _, s := range c.students {
		if _, ok := skip[s.ID]; ok {
			continue
		}
		return s, true
	}
	return c.students[0], true
}

// ptrTime 工具函数。
func ptrTime(t time.Time) *time.Time { return &t }
func ptrInt64(v int64) *int64        { return &v }
func ptrInt(v int) *int              { return &v }
func ptrFloat(v float64) *float64    { return &v }
func ptrStr(s string) *string        { return &s }

// encryptIDCard 加密身份证（AES-256-GCM）+ 计算 hash（用于去重）。
func encryptIDCard(idCard string) (string, string) {
	return cryptox.Encrypt(idCard), fmt.Sprintf("%x", sha256.Sum256([]byte(idCard)))
}

// nextBizNo 安全包装 idgen。
func nextBizNo(db *gorm.DB, module string) string {
	biz, err := idgen.NextBizNo(db, module)
	if err != nil {
		log.Fatalf("生成业务编号 %s 失败: %v", module, err)
	}
	return biz
}

// syncBizSeq 将 biz_seq.counter 与各业务表现存最大编号对齐，
// 避免手工灌入的存量数据与自增计数器脱节。
func (c *ctx) syncBizSeq() {
	year := c.now.Year()
	type spec struct {
		module string
		table  string
		col    string
	}
	specs := []spec{
		{"TY", "ty_application", "biz_no"},
		{"TY", "ty_recommendation_meeting", "biz_no"},
		{"TY", "ty_cultivation_record", "biz_no"},
		{"TY", "ty_development_object", "biz_no"},
		{"TY", "ty_development_meeting", "biz_no"},
		{"TY", "ty_probationary_meeting", "biz_no"},
		{"TY", "ty_member_roster", "biz_no"},
		{"ST", "st_association", "biz_no"},
		{"ST", "st_activity", "biz_no"},
		{"ST", "st_expense", "biz_no"},
		{"ST", "st_election", "biz_no"},
		{"SQ", "sq_inspection", "biz_no"},
		{"SQ", "sq_incident", "biz_no"},
		{"SQ", "sq_assessment", "biz_no"},
		{"QG", "qg_position", "biz_no"},
		{"QG", "qg_attendance", "biz_no"},
		{"QG", "qg_payroll", "biz_no"},
	}
	modMax := map[string]int{}
	for _, sp := range specs {
		var raw string
		row := c.db.Raw(
			"SELECT biz_no FROM "+sp.table+" WHERE biz_no LIKE ? AND is_deleted=0 ORDER BY biz_no DESC LIMIT 1",
			fmt.Sprintf("%s-%d-%%", sp.module, year),
		).Row()
		if err := row.Scan(&raw); err != nil {
			continue
		}
		// 解析末尾 4 位流水
		var seq int
		if _, err := fmt.Sscanf(raw, fmt.Sprintf("%s-%%d-%%d", sp.module), new(int), &seq); err != nil {
			continue
		}
		if seq > modMax[sp.module] {
			modMax[sp.module] = seq
		}
	}
	for module, maxSeq := range modMax {
		var cur int
		c.db.Raw("SELECT cur FROM biz_seq WHERE module=? AND year=?", module, year).Row().Scan(&cur)
		if maxSeq > cur {
			c.db.Exec(
				"UPDATE biz_seq SET cur=? WHERE module=? AND year=?",
				maxSeq, module, year,
			)
			fmt.Printf("  [sync] %s-%d 计数器 %d → %d\n", module, year, cur, maxSeq)
		}
	}
}

// count 简单包装。对有 is_deleted 的业务表做软删过滤；
// 若该表没有 is_deleted 列，则自动回退到全量统计。
func count(db *gorm.DB, model interface{}) int64 {
	var n int64
	err := db.Model(model).Where("is_deleted = 0").Count(&n).Error
	if err != nil && isMissingIsDeletedCol(err) {
		return countNoDel(db, model)
	}
	return n
}

// countNoDel 对没有 is_deleted 列的子表（签到 / 考勤等）做总数统计。
func countNoDel(db *gorm.DB, model interface{}) int64 {
	var n int64
	if err := db.Model(model).Count(&n).Error; err != nil {
		return 0
	}
	return n
}

// isMissingIsDeletedCol 判断错误是否为 "no such column: is_deleted"。
func isMissingIsDeletedCol(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "no such column: is_deleted")
}

// 通用：5 态状态机状态常量。
const (
	S0 = "S0" // 草稿
	S1 = "S1" // 待审 / 已提交
	S2 = "S2" // 审批中（中间态）
	S3 = "S3" // 通过 / 完成
	S4 = "S4" // 驳回 / 关闭
)

// printSummary 输出 4 大模块的关键表当前记录数。
func (c *ctx) printSummary() {
	tables := []struct {
		Module string
		Name   string
		Model  interface{}
	}{
		// TY
		{"TY", "ty_branch", &models.TyBranch{}},
		{"TY", "ty_application", &models.TyApplication{}},
		{"TY", "ty_approval_record", &models.TyApprovalRecord{}},
		{"TY", "ty_recommendation_meeting", &models.TyRecommendationMeeting{}},
		{"TY", "ty_recommendation_vote", &models.TyRecommendationVote{}},
		{"TY", "ty_cultivation_link", &models.TyCultivationLink{}},
		{"TY", "ty_cultivation_record", &models.TyCultivationRecord{}},
		{"TY", "ty_course_record", &models.TyCourseRecord{}},
		{"TY", "ty_thought_report", &models.TyThoughtReport{}},
		{"TY", "ty_development_object", &models.TyDevelopmentObject{}},
		{"TY", "ty_political_review", &models.TyPoliticalReview{}},
		{"TY", "ty_development_meeting", &models.TyDevelopmentMeeting{}},
		{"TY", "ty_probationary_record", &models.TyProbationaryRecord{}},
		{"TY", "ty_probationary_meeting", &models.TyProbationaryMeeting{}},
		{"TY", "ty_member_roster", &models.TyMemberRoster{}},
		// ST
		{"ST", "st_association", &models.StAssociation{}},
		{"ST", "st_activity", &models.StActivity{}},
		{"ST", "st_activity_approval", &models.StActivityApproval{}},
		{"ST", "st_activity_checkin", &models.StActivityCheckin{}},
		{"ST", "st_activity_summary", &models.StActivitySummary{}},
		{"ST", "st_expense", &models.StExpense{}},
		{"ST", "st_rating", &models.StRating{}},
		// SQ
		{"SQ", "sq_selfgov_position", &models.SqSelfgovPosition{}},
		{"SQ", "sq_inspection", &models.SqInspection{}},
		{"SQ", "sq_incident", &models.SqIncident{}},
		{"SQ", "sq_late_return", &models.SqLateReturn{}},
		{"SQ", "sq_violation", &models.SqViolation{}},
		{"SQ", "sq_assessment", &models.SqAssessment{}},
		// QG
		{"QG", "qg_difficulty_cert", &models.QgDifficultyCert{}},
		{"QG", "qg_position", &models.QgPosition{}},
		{"QG", "qg_position_apply", &models.QgPositionApply{}},
		{"QG", "qg_attendance", &models.QgAttendance{}},
		{"QG", "qg_monthly_assess", &models.QgMonthlyAssess{}},
		{"QG", "qg_payroll", &models.QgPayroll{}},
	}
	curModule := ""
	for _, t := range tables {
		if t.Module != curModule {
			fmt.Printf("\n[%s]\n", t.Module)
			curModule = t.Module
		}
		n := count(c.db, t.Model)
		fmt.Printf("  %-30s %4d\n", t.Name, n)
	}
}
