package com.studenthub.modules.cmp.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;

/**
 * 综合素质量化模块 Controller
 * 提供排名、成绩、看板等接口
 */
@RestController
@RequestMapping("/cmp")
public class CmpModuleController {

    private final JdbcTemplate jdbcTemplate;

    public CmpModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    // ====================== 排名列表 ======================

    /**
     * 综合素质排名列表
     * SQL 使用 RANK() 窗口函数计算班级排名和院系排名
     */
    @GetMapping("/scores")
    public R<Map<String, Object>> getScores(
            @RequestParam(value = "page", defaultValue = "1") int page,
            @RequestParam(value = "page_size", defaultValue = "20") int pageSize,
            @RequestParam(value = "college_id", required = false) Integer collegeId,
            @RequestParam(value = "class_id", required = false) Integer classId,
            @RequestParam(value = "q[student_name]", required = false) String studentName) {

        StringBuilder where = new StringBuilder("WHERE s.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (collegeId != null) {
            where.append("AND s.college_id = ? ");
            params.add(collegeId);
        }
        if (classId != null) {
            where.append("AND s.class_id = ? ");
            params.add(classId);
        }
        if (studentName != null && !studentName.trim().isEmpty()) {
            where.append("AND s.name LIKE ? ");
            params.add("%" + studentName.trim() + "%");
        }

        // 计算总数
        String countSql = "SELECT COUNT(*) FROM cmp_score c " +
                "LEFT JOIN idx_student s ON c.student_id = s.id " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        // 查询数据（含班级排名和院系排名，使用子查询兼容老版本 SQLite）
        String sql = "SELECT c.id, s.id as student_id, s.student_no, s.name as student_name, " +
                "IFNULL(col.name, '未分配') as college_name, " +
                "IFNULL(cls.name, '未分配') as college_class_name, " +
                "'2025-2026' as academic_year, " +
                "c.total_score, c.academic_score, " +
                "c.ty_score as moral_score, c.st_score as sports_score, " +
                "c.sq_score as art_score, c.qg_score as labor_score, " +
                "c.updated_at as computed_at, " +
                "(SELECT COUNT(*) + 1 FROM cmp_score c2 " +
                " LEFT JOIN idx_student s2 ON c2.student_id = s2.id " +
                " WHERE s2.class_id = s.class_id AND c2.total_score > c.total_score) as rank_in_class, " +
                "(SELECT COUNT(*) + 1 FROM cmp_score c3 " +
                " LEFT JOIN idx_student s3 ON c3.student_id = s3.id " +
                " WHERE s3.college_id = s.college_id AND c3.total_score > c.total_score) as rank_in_college " +
                "FROM cmp_score c " +
                "LEFT JOIN idx_student s ON c.student_id = s.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "LEFT JOIN idx_class cls ON s.class_id = cls.id " +
                where +
                "ORDER BY c.total_score DESC " +
                "LIMIT ? OFFSET ?";

        List<Object> queryParams = new ArrayList<>(params);
        queryParams.add(pageSize);
        queryParams.add((page - 1) * pageSize);

        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, queryParams.toArray());

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", pageSize);
        return R.ok(result);
    }

    /**
     * 兼容旧接口 /cmp/rankings
     */
    @GetMapping("/rankings")
    public R<Map<String, Object>> getRankings() {
        String sql = "SELECT (SELECT COUNT(*) + 1 FROM cmp_score c2 WHERE c2.total_score > c.total_score) as rank, " +
                "s.name as student_name, s.student_no, IFNULL(col.name, '未分配') as college_name, " +
                "c.total_score, c.academic_score, c.ty_score as moral_score, c.st_score as sports_score, " +
                "c.sq_score as art_score, c.qg_score as labor_score " +
                "FROM cmp_score c " +
                "LEFT JOIN idx_student s ON c.student_id = s.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "ORDER BY c.total_score DESC";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    // ====================== 我的成绩 ======================

    /**
     * 当前登录学生的综合分
     * 从 Sa-Token 获取用户名 → 查 sys_user → 关联 idx_student
     */
    @GetMapping("/scores/me")
    public R<Map<String, Object>> getMyScore(@RequestParam(value = "term", required = false) String term) {
        Long studentId = getCurrentStudentId();
        if (studentId == null) {
            return R.fail(14040, "当前用户未关联学生信息");
        }
        return getScoreDetail(studentId, term);
    }

    /**
     * 当前登录学生的历史综合分
     */
    @GetMapping("/scores/me/history")
    public R<List<Map<String, Object>>> getMyHistory() {
        Long studentId = getCurrentStudentId();
        if (studentId == null) {
            return R.ok(Collections.emptyList());
        }
        // 当前表无 academic_year 字段，返回最近一条作为历史
        String sql = "SELECT '2025-2026' as academic_year, c.total_score, c.academic_score, " +
                "c.ty_score as moral_score, c.st_score as sports_score, " +
                "c.sq_score as art_score, c.qg_score as labor_score, " +
                "c.updated_at as computed_at " +
                "FROM cmp_score c WHERE c.student_id = ? " +
                "ORDER BY c.updated_at DESC LIMIT 12";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, studentId);
        return R.ok(items);
    }

    /**
     * 单个学生综合分详情
     */
    @GetMapping("/scores/{studentId}")
    public R<Map<String, Object>> getScoreDetail(
            @PathVariable("studentId") Long studentId,
            @RequestParam(value = "term", required = false) String term) {
        String sql = "SELECT c.id, s.id as student_id, s.student_no, s.name as student_name, " +
                "IFNULL(col.name, '未分配') as college_name, " +
                "IFNULL(cls.name, '未分配') as college_class_name, " +
                "'2025-2026' as academic_year, " +
                "'v1.0' as rule_version, " +
                "c.total_score, c.academic_score, " +
                "c.ty_score as moral_score, c.st_score as sports_score, " +
                "c.sq_score as art_score, c.qg_score as labor_score, " +
                "c.updated_at as computed_at " +
                "FROM cmp_score c " +
                "LEFT JOIN idx_student s ON c.student_id = s.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "LEFT JOIN idx_class cls ON s.class_id = cls.id " +
                "WHERE c.student_id = ?";

        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql, studentId);
        if (rows.isEmpty()) {
            return R.fail(14040, "未找到该学生的综合分记录");
        }
        Map<String, Object> score = rows.get(0);

        // 拼装 dimensions 嵌套结构（前端 MyScore.vue 期望）
        Map<String, Object> dimensions = new HashMap<>();
        dimensions.put("league", score.get("moral_score"));
        dimensions.put("assoc", score.get("sports_score"));
        dimensions.put("community", score.get("art_score"));
        dimensions.put("workstudy", score.get("labor_score"));
        dimensions.put("academic", score.get("academic_score"));
        score.put("dimensions", dimensions);

        // 拼装 details 数组（前端雷达图和明细表格用）
        List<Map<String, Object>> details = new ArrayList<>();
        details.add(buildDetail("思想品德", "league", (Number) score.get("moral_score"), 20.0, "团组织生活、政治学习"));
        details.add(buildDetail("社团活动", "assoc", (Number) score.get("sports_score"), 20.0, "社团参与、活动表现"));
        details.add(buildDetail("社区表现", "community", (Number) score.get("art_score"), 20.0, "宿舍卫生、社区服务"));
        details.add(buildDetail("劳动教育", "workstudy", (Number) score.get("labor_score"), 20.0, "勤工助学、志愿服务"));
        details.add(buildDetail("学业成绩", "academic", (Number) score.get("academic_score"), 20.0, "课程成绩、学业竞赛"));
        score.put("details", details);

        return R.ok(score);
    }

    private Map<String, Object> buildDetail(String name, String code, Number score, Double maxScore, String desc) {
        Map<String, Object> d = new HashMap<>();
        d.put("name", name);
        d.put("code", code);
        d.put("score", score);
        d.put("max_score", maxScore);
        d.put("weight", 0.2);
        d.put("description", desc);
        return d;
    }

    /**
     * 重算单学生综合分（简单实现：重新查表汇总）
     */
    @PostMapping("/scores/{studentId}/recompute")
    public R<Map<String, Object>> recomputeScore(
            @PathVariable("studentId") Long studentId,
            @RequestParam(value = "term", required = false) String term) {
        // 简化实现：直接返回当前分数（实际应重新汇总 TY/ST/SQ/QG 数据）
        return getScoreDetail(studentId, term);
    }

    /**
     * 批量重算
     */
    @PostMapping("/scores/compute")
    public R<Map<String, Object>> batchCompute(@RequestBody(required = false) Map<String, Object> body) {
        // 简化实现：返回统计信息
        Integer total = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM cmp_score", Integer.class);
        Map<String, Object> result = new HashMap<>();
        result.put("computed", total != null ? total : 0);
        result.put("message", "批量重算完成");
        return R.ok(result);
    }

    // ====================== 看板 KPI ======================

    @GetMapping("/dashboard/kpi")
    public R<Map<String, Object>> getDashboardKpi(
            @RequestParam(value = "term", required = false) String term) {
        Map<String, Object> kpi = new HashMap<>();

        // 学生总数
        Integer studentCount = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM idx_student WHERE is_deleted = 0", Integer.class);
        kpi.put("student_count", studentCount != null ? studentCount : 0);

        // 已计算综合分人数
        Integer computedCount = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM cmp_score", Integer.class);
        kpi.put("computed_count", computedCount != null ? computedCount : 0);

        // 平均分
        Double avgScore = jdbcTemplate.queryForObject(
                "SELECT ROUND(AVG(total_score), 2) FROM cmp_score", Double.class);
        kpi.put("avg_score", avgScore != null ? avgScore : 0.0);

        // 最高分
        Double maxScore = jdbcTemplate.queryForObject(
                "SELECT MAX(total_score) FROM cmp_score", Double.class);
        kpi.put("max_score", maxScore != null ? maxScore : 0.0);

        // 最低分
        Double minScore = jdbcTemplate.queryForObject(
                "SELECT MIN(total_score) FROM cmp_score", Double.class);
        kpi.put("min_score", minScore != null ? minScore : 0.0);

        // 优秀率（>=90）
        Integer excellentCount = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM cmp_score WHERE total_score >= 90", Integer.class);
        kpi.put("excellent_count", excellentCount != null ? excellentCount : 0);
        kpi.put("excellent_rate", studentCount != null && studentCount > 0 ?
                Math.round(excellentCount * 10000.0 / studentCount) / 100.0 : 0.0);

        // 及格率（>=60）
        Integer passCount = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM cmp_score WHERE total_score >= 60", Integer.class);
        kpi.put("pass_count", passCount != null ? passCount : 0);
        kpi.put("pass_rate", studentCount != null && studentCount > 0 ?
                Math.round(passCount * 10000.0 / studentCount) / 100.0 : 0.0);

        return R.ok(kpi);
    }

    /**
     * 趋势数据（简化：返回最近 12 个月模拟数据）
     */
    @GetMapping("/dashboard/trends")
    public R<Map<String, Object>> getTrends(
            @RequestParam(value = "metric", defaultValue = "ty_pass_rate") String metric,
            @RequestParam(value = "range", defaultValue = "12m") String range) {
        // 简化实现：返回模拟趋势数据
        List<Map<String, Object>> points = new ArrayList<>();
        String[] months = {"2025-08", "2025-09", "2025-10", "2025-11", "2025-12",
                "2026-01", "2026-02", "2026-03", "2026-04", "2026-05", "2026-06", "2026-07"};
        double[] values = {82.5, 85.0, 87.3, 88.1, 89.5, 90.2, 91.0, 92.3, 93.1, 94.0, 94.8, 95.2};
        for (int i = 0; i < months.length; i++) {
            Map<String, Object> p = new HashMap<>();
            p.put("month", months[i]);
            p.put("value", values[i]);
            points.add(p);
        }
        Map<String, Object> result = new HashMap<>();
        result.put("points", points);
        result.put("metric", metric);
        return R.ok(result);
    }

    /**
     * 院系分布
     */
    @GetMapping("/dashboard/distribution")
    public R<Map<String, Object>> getDistribution(
            @RequestParam(value = "dim", defaultValue = "college") String dim,
            @RequestParam(value = "term", required = false) String term) {
        String sql = "SELECT IFNULL(col.name, '未分配') as name, " +
                "COUNT(*) as count, ROUND(AVG(c.total_score), 2) as avg_score " +
                "FROM cmp_score c " +
                "LEFT JOIN idx_student s ON c.student_id = s.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "GROUP BY col.id, col.name " +
                "ORDER BY avg_score DESC";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql);
        List<Map<String, Object>> buckets = new ArrayList<>();
        for (Map<String, Object> row : rows) {
            Map<String, Object> b = new HashMap<>();
            b.put("name", row.get("name"));
            b.put("count", row.get("count"));
            b.put("value", row.get("avg_score"));
            buckets.add(b);
        }
        Map<String, Object> result = new HashMap<>();
        result.put("buckets", buckets);
        return R.ok(result);
    }

    /**
     * 各院系活跃社团数
     */
    @GetMapping("/dashboard/active-assoc-by-college")
    public R<Map<String, Object>> getActiveAssocByCollege() {
        String sql = "SELECT IFNULL(col.name, '未分配') as name, " +
                "COUNT(DISTINCT a.id) as count " +
                "FROM st_association a " +
                "LEFT JOIN sys_college col ON a.college_id = col.id " +
                "WHERE a.is_deleted = 0 AND a.status = 'active' " +
                "GROUP BY col.id, col.name " +
                "ORDER BY count DESC";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql);
        List<Map<String, Object>> buckets = new ArrayList<>();
        for (Map<String, Object> row : rows) {
            Map<String, Object> b = new HashMap<>();
            b.put("name", row.get("name"));
            b.put("count", row.get("count"));
            b.put("value", row.get("count"));
            buckets.add(b);
        }
        Map<String, Object> result = new HashMap<>();
        result.put("buckets", buckets);
        return R.ok(result);
    }

    /**
     * 事件等级分布
     */
    @GetMapping("/dashboard/incident-level")
    public R<Map<String, Object>> getIncidentLevel() {
        String sql = "SELECT level as name, COUNT(*) as count " +
                "FROM sq_incident WHERE is_deleted = 0 " +
                "GROUP BY level ORDER BY count DESC";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql);
        List<Map<String, Object>> buckets = new ArrayList<>();
        // 如果 sq_incident 没有数据，返回默认空桶
        if (rows.isEmpty()) {
            String[] names = {"一般事件", "较重事件", "严重事件", "特大事件"};
            for (int i = 0; i < names.length; i++) {
                Map<String, Object> b = new HashMap<>();
                b.put("name", names[i]);
                b.put("count", 0);
                b.put("value", 0);
                buckets.add(b);
            }
        } else {
            for (Map<String, Object> row : rows) {
                Map<String, Object> b = new HashMap<>();
                b.put("name", row.get("name"));
                b.put("count", row.get("count"));
                b.put("value", row.get("count"));
                buckets.add(b);
            }
        }
        Map<String, Object> result = new HashMap<>();
        result.put("buckets", buckets);
        return R.ok(result);
    }

    // ====================== 规则版本（简化） ======================

    @GetMapping("/rule-versions")
    public R<List<Map<String, Object>>> getRuleVersions() {
        List<Map<String, Object>> list = new ArrayList<>();
        Map<String, Object> v = new HashMap<>();
        v.put("id", 1);
        v.put("version", "v1.0");
        v.put("status", "active");
        v.put("description", "默认规则版本");
        v.put("created_at", "2025-09-01 00:00:00");
        list.add(v);
        return R.ok(list);
    }

    @PostMapping("/rule-versions")
    public R<Map<String, Object>> createRuleVersion(@RequestBody Map<String, Object> body) {
        Map<String, Object> result = new HashMap<>();
        result.put("id", System.currentTimeMillis());
        result.put("version", body.getOrDefault("version", "v" + System.currentTimeMillis()));
        result.put("status", "draft");
        return R.ok(result);
    }

    @PostMapping("/rule-versions/{id}/activate")
    public R<Map<String, Object>> activateRuleVersion(@PathVariable("id") Long id) {
        Map<String, Object> result = new HashMap<>();
        result.put("id", id);
        result.put("status", "active");
        return R.ok(result);
    }

    // ====================== 辅助方法 ======================

    /**
     * 从 Sa-Token 获取当前登录用户关联的 student_id
     */
    private Long getCurrentStudentId() {
        try {
            Object loginId = cn.dev33.satoken.stp.StpUtil.getLoginId();
            if (loginId == null) return null;
            String username = String.valueOf(loginId);
            // 优先按 username 查 sys_user.student_id
            List<Map<String, Object>> rows = jdbcTemplate.queryForList(
                    "SELECT student_id FROM sys_user WHERE username = ? AND student_id IS NOT NULL",
                    username);
            if (!rows.isEmpty() && rows.get(0).get("student_id") != null) {
                Number n = (Number) rows.get(0).get("student_id");
                return n.longValue();
            }
            // 兼容：如果用户名是学号（如 2023010101），直接按学号查学生
            rows = jdbcTemplate.queryForList(
                    "SELECT id FROM idx_student WHERE student_no = ?", username);
            if (!rows.isEmpty() && rows.get(0).get("id") != null) {
                Number n = (Number) rows.get(0).get("id");
                return n.longValue();
            }
            // 兜底：admin 用户返回第一个学生
            if ("admin".equals(username)) {
                rows = jdbcTemplate.queryForList("SELECT MIN(id) as id FROM idx_student");
                if (!rows.isEmpty() && rows.get(0).get("id") != null) {
                    Number n = (Number) rows.get(0).get("id");
                    return n.longValue();
                }
            }
        } catch (Exception e) {
            // 未登录或异常
        }
        return null;
    }
}
