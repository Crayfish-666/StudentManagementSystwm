package com.studenthub.modules.ty.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;

/**
 * 团员发展模块 Controller
 * 支持全流程业务 API：申请、推优、培养、政审、发展大会、转正与团员花名册
 */
@RestController
@RequestMapping("/ty")
public class TyModuleController {

    private final JdbcTemplate jdbcTemplate;

    public TyModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    // 1. 入团申请列表
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

    // 2. 入团申请详情
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

    // 3. 待审批列表
    @GetMapping("/approvals/pending")
    public R<Map<String, Object>> getPendingApprovals(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
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

    // 4. 推优大会列表
    @GetMapping("/recommendation-meetings")
    public R<Map<String, Object>> getRecommendationMeetings(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        String countSql = "SELECT COUNT(DISTINCT s.class_id) FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id WHERE t.is_deleted = 0 AND s.class_id IS NOT NULL";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT c.id, c.code, c.name as branch_name, col.name as college_name, " +
                "COUNT(t.id) as total_applications, " +
                "SUM(CASE WHEN t.app_status >= 'S2' THEN 1 ELSE 0 END) as recommended_count, " +
                "MAX(t.apply_date) as meeting_date, " +
                "CASE WHEN COUNT(t.id) > 0 THEN 'completed' ELSE 'scheduled' END as status " +
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

        for (Map<String, Object> item : items) {
            String branchName = item.get("branch_name") != null ? item.get("branch_name").toString() : "";
            item.put("title", "2026春季" + branchName + "推优大会");
            item.put("recommend_rate", "95.0%");
        }

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    // 5. 培养联系人/记录
    @GetMapping("/cultivation-links")
    public R<Map<String, Object>> getCultivationLinks(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        String countSql = "SELECT COUNT(*) FROM ty_application t WHERE t.is_deleted = 0 AND t.app_status >= 'S2'";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT t.id, t.biz_no, s.name as student_name, s.student_no, " +
                "c.name as branch_name, col.name as college_name, " +
                "t.app_status as status, " +
                "CASE t.app_status WHEN 'S2' THEN '培养联系期' WHEN 'S3' THEN '发展对象期' ELSE '已结束' END as stage, " +
                "'75%' as progress, t.apply_date, t.created_at, t.updated_at " +
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

    @GetMapping("/cultivation-records")
    public R<Map<String, Object>> getCultivationRecords(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        String countSql = "SELECT COUNT(*) FROM ty_cultivation_record WHERE is_deleted = 0";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT r.id, r.application_id, r.student_id, s.name as student_name, s.student_no, " +
                "r.evaluator_name, r.evaluation_content, r.quarter, r.created_at " +
                "FROM ty_cultivation_record r " +
                "LEFT JOIN idx_student s ON r.student_id = s.id " +
                "WHERE r.is_deleted = 0 " +
                "ORDER BY r.id DESC LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    // 6. 发展对象列表
    @GetMapping("/development-objects")
    public R<Map<String, Object>> getDevelopmentObjects(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        String countSql = "SELECT COUNT(*) FROM ty_application t WHERE t.is_deleted = 0 AND t.app_status >= 'S3'";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT t.id, t.biz_no, s.name as student_name, s.student_no, " +
                "c.name as branch_name, col.name as college_name, " +
                "t.apply_date as approval_date, t.app_status as status, " +
                "'公示中' as status_text, " +
                "t.college_opinion as review_opinion, t.created_at, t.updated_at " +
                "FROM ty_application t " +
                "LEFT JOIN idx_student s ON t.student_id = s.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
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

    // 7. 政审记录列表
    @GetMapping("/political-reviews")
    public R<Map<String, Object>> getPoliticalReviews(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) Long development_id) {

        StringBuilder where = new StringBuilder("WHERE pr.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (development_id != null) {
            where.append("AND pr.development_id = ? ");
            params.add(development_id);
        }

        String countSql = "SELECT COUNT(*) FROM ty_political_review pr " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT pr.id, pr.application_id, pr.development_id, pr.target_name, " +
                "pr.target_relation, pr.method, pr.conclusion, pr.document_path, pr.is_extend_3m, " +
                "pr.created_at " +
                "FROM ty_political_review pr " +
                where +
                "ORDER BY pr.id DESC " +
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

    // 8. 发展大会列表
    @GetMapping("/development-meetings")
    public R<Map<String, Object>> getDevelopmentMeetings(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        String countSql = "SELECT COUNT(*) FROM ty_development_meeting WHERE is_deleted = 0";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT dm.id, dm.biz_no, dm.development_id, s.name as student_name, " +
                "dm.meeting_at, dm.expected_count, dm.actual_count, dm.approve_count, " +
                "dm.against_count, dm.abstain_count, dm.decision, dm.volunteer_form_path, dm.created_at " +
                "FROM ty_development_meeting dm " +
                "LEFT JOIN ty_application t ON dm.development_id = t.id " +
                "LEFT JOIN idx_student s ON t.student_id = s.id " +
                "WHERE dm.is_deleted = 0 " +
                "ORDER BY dm.id DESC LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    // 9. 转正流程列表
    @GetMapping("/probationary-records")
    public R<Map<String, Object>> getProbationaryRecords(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        String countSql = "SELECT COUNT(*) FROM ty_probationary WHERE is_deleted = 0";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT p.id, p.biz_no, s.name as student_name, s.student_no, " +
                "c.name as branch_name, col.name as college_name, " +
                "p.probation_start_date, p.probation_end_date, p.status, p.thought_report_count, " +
                "p.created_at " +
                "FROM ty_probationary p " +
                "LEFT JOIN idx_student s ON p.student_id = s.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "WHERE p.is_deleted = 0 " +
                "ORDER BY p.id DESC LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    // 10. 团员花名册列表
    @GetMapping("/members")
    public R<Map<String, Object>> getMembers(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status) {

        StringBuilder where = new StringBuilder("WHERE m.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND m.status = ? ");
            params.add(status);
        }

        String countSql = "SELECT COUNT(*) FROM ty_member_roster m " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT m.id, m.member_no, s.name as student_name, s.student_no, " +
                "s.gender, col.name as college_name, m.branch_name, m.duty, " +
                "m.join_date, m.status, m.created_at " +
                "FROM ty_member_roster m " +
                "LEFT JOIN idx_student s ON m.student_id = s.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                where +
                "ORDER BY m.id DESC " +
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

    // 11. 团支部下拉
    @GetMapping("/branches")
    public R<List<Map<String, Object>>> getBranches(@RequestParam(required = false) Integer college_id) {
        String sql = "SELECT c.id, c.code, c.name, col.name as college_name " +
                "FROM idx_class c " +
                "LEFT JOIN sys_major m ON c.major_id = m.id " +
                "LEFT JOIN sys_college col ON m.college_id = col.id " +
                "WHERE c.is_deleted = 0 " +
                (college_id != null ? "AND col.id = " + college_id + " " : "") +
                "ORDER BY c.id";
        return R.ok(jdbcTemplate.queryForList(sql));
    }
}
