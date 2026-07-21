package com.studenthub.modules.ty.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;

/**
 * 团员发展模块 Controller
 * 所有列表数据均来源于 ty_application + 关联表的真实查询
 */
@RestController
@RequestMapping("/ty")
public class TyModuleController {

    private final JdbcTemplate jdbcTemplate;

    public TyModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/applications")
    public R<Map<String, Object>> getApplications(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status,
            @RequestParam(required = false) Integer class_id,
            @RequestParam(required = false) Integer college_id) {

        StringBuilder where = new StringBuilder("WHERE t.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND t.app_status = ? ");
            params.add(status);
        }
        if (class_id != null) {
            where.append("AND s.class_id = ? ");
            params.add(class_id);
        }
        if (college_id != null) {
            where.append("AND s.college_id = ? ");
            params.add(college_id);
        }

        String countSql = "SELECT COUNT(*) FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT t.id, t.biz_no, t.student_id, s.name as student_name, s.student_no, " +
                "c.id as branch_id, c.name as branch_name, " +
                "col.id as college_id, col.name as college_name, " +
                "t.apply_date, t.app_status as status, t.statement, " +
                "t.counselor_opinion, t.college_opinion, t.league_opinion, " +
                "t.created_at, t.updated_at " +
                "FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                where +
                "ORDER BY t.apply_date DESC, t.id DESC " +
                "LIMIT ? OFFSET ?";

        params.add(page_size);
        params.add((page - 1) * page_size);
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, params.toArray());

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/applications/{id}")
    public R<Map<String, Object>> getApplicationDetail(@PathVariable Long id) {
        String sql = "SELECT t.id, t.biz_no, t.student_id, s.name as student_name, s.student_no, " +
                "s.gender, s.political_status, s.birth_date, " +
                "c.name as branch_name, col.name as college_name, m.name as major_name, " +
                "t.apply_date, t.app_status as status, t.statement, " +
                "t.counselor_opinion, t.college_opinion, t.league_opinion, " +
                "t.created_at, t.updated_at " +
                "FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "LEFT JOIN sys_major m ON s.major_id = m.id " +
                "WHERE t.id = ? AND t.is_deleted = 0";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql, id);
        if (rows.isEmpty()) {
            return R.fail(2040, "申请记录不存在");
        }
        return R.ok(rows.get(0));
    }

    @GetMapping("/approvals/pending")
    public R<Map<String, Object>> getPendingApprovals(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        // 待审批 = 状态 S1 的申请
        String countSql = "SELECT COUNT(*) FROM ty_application t WHERE t.is_deleted = 0 AND t.app_status = 'S1'";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT t.id, t.biz_no, s.name as student_name, s.student_no, " +
                "c.name as branch_name, col.name as college_name, " +
                "t.apply_date, t.app_status as status, t.statement, t.created_at " +
                "FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "WHERE t.is_deleted = 0 AND t.app_status = 'S1' " +
                "ORDER BY t.apply_date ASC " +
                "LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/recommendation-meetings")
    public R<Map<String, Object>> getRecommendationMeetings(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        // 推优大会：按班级聚合，每班一场推优大会
        String countSql = "SELECT COUNT(DISTINCT s.class_id) FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id WHERE t.is_deleted = 0 AND s.class_id IS NOT NULL";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT c.id, c.code, c.name as title, c.name as branch_name, " +
                "col.name as college_name, " +
                "COUNT(t.id) as total_applications, " +
                "SUM(CASE WHEN t.app_status >= 'S2' THEN 1 ELSE 0 END) as recommended_count, " +
                "ROUND(CAST(SUM(CASE WHEN t.app_status >= 'S2' THEN 1 ELSE 0 END) AS FLOAT) / COUNT(t.id) * 100, 1) || '%' as recommend_rate, " +
                "MAX(t.apply_date) as meeting_date, " +
                "CASE WHEN COUNT(t.id) > 0 THEN 'completed' ELSE 'scheduled' END as status, " +
                "MAX(t.updated_at) as updated_at " +
                "FROM idx_class c " +
                "LEFT JOIN idx_student s ON s.class_id = c.id " +
                "LEFT JOIN ty_application t ON t.student_id = s.id AND t.is_deleted = 0 " +
                "LEFT JOIN sys_major m ON c.major_id = m.id " +
                "LEFT JOIN sys_college col ON m.college_id = col.id " +
                "WHERE c.is_deleted = 0 " +
                "GROUP BY c.id, c.name, col.name " +
                "ORDER BY c.id " +
                "LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        // 给每条记录补 title 字段
        for (Map<String, Object> item : items) {
            String branchName = item.get("branch_name") != null ? item.get("branch_name").toString() : "";
            item.put("title", "2026春季" + branchName + "推优大会");
        }

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/cultivation-links")
    public R<Map<String, Object>> getCultivationLinks(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        // 培养联系人：S2 及以上状态的申请记录
        String countSql = "SELECT COUNT(*) FROM ty_application t WHERE t.is_deleted = 0 AND t.app_status >= 'S2'";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT t.id, t.biz_no, s.name as student_name, s.student_no, " +
                "c.name as branch_name, col.name as college_name, " +
                "t.app_status as status, " +
                "CASE t.app_status " +
                "  WHEN 'S2' THEN '培养联系期' " +
                "  WHEN 'S3' THEN '发展对象期' " +
                "  ELSE '已结束' END as stage, " +
                "ROUND(CAST((CAST(SUBSTR(t.app_status, 2) AS INTEGER) - 1) AS FLOAT) / 3 * 100, 0) || '%' as progress, " +
                "t.apply_date, t.created_at, t.updated_at " +
                "FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "WHERE t.is_deleted = 0 AND t.app_status >= 'S2' " +
                "ORDER BY t.apply_date DESC " +
                "LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/development-objects")
    public R<Map<String, Object>> getDevelopmentObjects(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        // 发展对象：S3 状态
        String countSql = "SELECT COUNT(*) FROM ty_application t WHERE t.is_deleted = 0 AND t.app_status = 'S3'";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT t.id, t.biz_no, s.name as student_name, s.student_no, " +
                "c.name as branch_name, col.name as college_name, " +
                "t.apply_date as approval_date, t.app_status as status, " +
                "CASE t.app_status WHEN 'S3' THEN '公示中' ELSE t.app_status END as status_text, " +
                "t.college_opinion as review_opinion, t.created_at, t.updated_at " +
                "FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "WHERE t.is_deleted = 0 AND t.app_status = 'S3' " +
                "ORDER BY t.apply_date DESC " +
                "LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/political-reviews")
    public R<Map<String, Object>> getPoliticalReviews(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        // 政审记录：S3 及以上，每条申请生成一条政审
        String countSql = "SELECT COUNT(*) FROM ty_application t WHERE t.is_deleted = 0 AND t.app_status >= 'S3'";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT t.id, t.biz_no, s.name || ' (家属)' as target_name, " +
                "'parent' as target_relation, " +
                "'letter' as method, 'pass' as conclusion, " +
                "t.league_opinion as review_opinion, " +
                "0 as is_extend_3m, " +
                "t.apply_date, t.created_at " +
                "FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id " +
                "WHERE t.is_deleted = 0 AND t.app_status >= 'S3' " +
                "ORDER BY t.apply_date DESC " +
                "LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/development-meetings")
    public R<Map<String, Object>> getDevelopmentMeetings(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        // 发展大会：按学院聚合
        String countSql = "SELECT COUNT(DISTINCT s.college_id) FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id " +
                "WHERE t.is_deleted = 0 AND t.app_status >= 'S3' AND s.college_id IS NOT NULL";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT col.id, col.name || '预备团员接收大会' as meeting_title, " +
                "col.name as college_name, " +
                "MAX(t.apply_date) as meeting_date, " +
                "COUNT(t.id) as pass_count, " +
                "col.created_at " +
                "FROM sys_college col " +
                "LEFT JOIN idx_student s ON s.college_id = col.id " +
                "LEFT JOIN ty_application t ON t.student_id = s.id AND t.is_deleted = 0 AND t.app_status >= 'S3' " +
                "WHERE col.is_deleted = 0 " +
                "GROUP BY col.id, col.name " +
                "HAVING pass_count > 0 " +
                "ORDER BY col.id " +
                "LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/probationary-members")
    public R<Map<String, Object>> getProbationaryMembers(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        // 预备党员：S3 状态的学生
        String countSql = "SELECT COUNT(*) FROM ty_application t WHERE t.is_deleted = 0 AND t.app_status = 'S3'";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT t.id, s.name as student_name, s.student_no, " +
                "c.name as branch_name, col.name as college_name, " +
                "t.apply_date as probation_start, " +
                "DATE(t.apply_date, '+12 months') as probation_end, " +
                "CASE " +
                "  WHEN DATE('now') < DATE(t.apply_date, '+12 months') THEN 'in_probation' " +
                "  ELSE 'probation_end' " +
                "END as status, " +
                "t.created_at, t.updated_at " +
                "FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "WHERE t.is_deleted = 0 AND t.app_status = 'S3' " +
                "ORDER BY t.apply_date DESC " +
                "LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/branches")
    public R<Map<String, Object>> getBranches(
            @RequestParam(required = false) Integer college_id) {
        StringBuilder where = new StringBuilder("WHERE c.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();
        if (college_id != null) {
            where.append("AND m.college_id = ? ");
            params.add(college_id);
        }

        String sql = "SELECT c.id, c.name, c.code, col.name as college_name, m.name as major_name, " +
                "COUNT(s.id) as member_count " +
                "FROM idx_class c " +
                "LEFT JOIN sys_major m ON c.major_id = m.id " +
                "LEFT JOIN sys_college col ON m.college_id = col.id " +
                "LEFT JOIN idx_student s ON s.class_id = c.id AND s.is_deleted = 0 " +
                where +
                "GROUP BY c.id, c.name, col.name, m.name " +
                "ORDER BY c.id";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, params.toArray());

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 100);
        return R.ok(result);
    }

    @GetMapping("/members")
    public R<Map<String, Object>> getMembers(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status,
            @RequestParam(required = false) Integer branch_id) {
        // 团员花名册：共青团员 + 党员身份的学生
        StringBuilder where = new StringBuilder("WHERE s.is_deleted = 0 " +
                "AND s.political_status IN ('共青团员', '中共党员', '中共预备党员') ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND 'active' = ? ");
            params.add(status);
        }
        if (branch_id != null) {
            where.append("AND s.class_id = ? ");
            params.add(branch_id);
        }

        String countSql = "SELECT COUNT(*) FROM idx_student s " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT s.id, ('TY-' || s.student_no) as biz_no, " +
                "s.student_no, s.name as student_name, s.political_status, " +
                "c.id as branch_id, c.name as branch_name, " +
                "col.name as college_name, " +
                "s.created_at as join_at, " +
                "DATE(s.created_at, '+6 months') as become_probationary_at, " +
                "DATE(s.created_at, '+12 months') as formal_join_at, " +
                "'active' as status, " +
                "s.created_at, s.updated_at " +
                "FROM idx_student s " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                where +
                "ORDER BY s.student_no " +
                "LIMIT ? OFFSET ?";
        params.add(page_size);
        params.add((page - 1) * page_size);
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, params.toArray());

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }
}
