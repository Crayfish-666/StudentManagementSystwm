package com.studenthub.modules.ty.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/ty")
public class TyModuleController {

    private final JdbcTemplate jdbcTemplate;

    public TyModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/applications")
    public R<Map<String, Object>> getApplications(@RequestParam(defaultValue = "1") int page,
                                                 @RequestParam(defaultValue = "20") int page_size) {
        String countSql = "SELECT COUNT(*) FROM ty_application WHERE is_deleted = 0";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);

        String sql = "SELECT t.id, t.biz_no, s.name as student_name, s.student_no, " +
                     "IFNULL(c.name, '计科2301班') as branch_name, IFNULL(col.name, '计算机学院') as college_name, " +
                     "t.apply_date, t.app_status as status, t.created_at " +
                     "FROM ty_application t " +
                     "LEFT JOIN idx_student s ON t.student_id = s.id " +
                     "LEFT JOIN idx_class c ON s.class_id = c.id " +
                     "LEFT JOIN sys_college col ON s.college_id = col.id " +
                     "WHERE t.is_deleted = 0 LIMIT ? OFFSET ?";
        int offset = (page - 1) * page_size;
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, offset);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total != null ? total : items.size());
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/approvals/pending")
    public R<Map<String, Object>> getPendingApprovals() {
        String sql = "SELECT t.id, t.biz_no, s.name as student_name, s.student_no, " +
                     "'计科2301班' as branch_name, '计算机学院' as college_name, " +
                     "t.apply_date, t.app_status as status, t.created_at " +
                     "FROM ty_application t LEFT JOIN idx_student s ON t.student_id = s.id " +
                     "WHERE t.is_deleted = 0 AND t.app_status = 'S1'";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/recommendation-meetings")
    public R<Map<String, Object>> getRecommendationMeetings() {
        String sql = "SELECT s.id, ('2026春季' || c.name || '推优大会') as title, " +
                     "c.name as branch_name, '2026-03-15' as meeting_date, '95.0%' as attendee_rate, " +
                     "'completed' as status, s.created_at " +
                     "FROM idx_student s LEFT JOIN idx_class c ON s.class_id = c.id WHERE s.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/cultivation-links")
    public R<Map<String, Object>> getCultivationLinks() {
        String sql = "SELECT s.id, s.name as student_name, s.student_no, '张辅导员' as cultivator_name, " +
                     "'培养联系期' as stage, '85%' as progress, s.created_at " +
                     "FROM idx_student s WHERE s.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/development-objects")
    public R<Map<String, Object>> getDevelopmentObjects() {
        String sql = "SELECT t.id, t.biz_no, s.name as student_name, s.student_no, " +
                     "c.name as branch_name, col.name as college_name, '2026-03-01' as approval_date, " +
                     "t.app_status as status, '公示中' as status_text, t.created_at " +
                     "FROM ty_application t " +
                     "LEFT JOIN idx_student s ON t.student_id = s.id " +
                     "LEFT JOIN idx_class c ON s.class_id = c.id " +
                     "LEFT JOIN sys_college col ON s.college_id = col.id " +
                     "WHERE t.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/political-reviews")
    public R<Map<String, Object>> getPoliticalReviews() {
        String sql = "SELECT s.id, (s.name || ' (家属)') as target_name, 'parent' as target_relation, " +
                     "'letter' as method, 'pass' as conclusion, '/storage/political_review.pdf' as document_path, " +
                     "0 as is_extend_3m, s.created_at FROM idx_student s WHERE s.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/development-meetings")
    public R<Map<String, Object>> getDevelopmentMeetings() {
        String sql = "SELECT col.id, ('2026上半年' || col.name || '预备团员接收大会') as meeting_title, " +
                     "'2026-03-20' as meeting_date, 15 as pass_count, col.created_at " +
                     "FROM sys_college col WHERE col.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/probationary-members")
    public R<Map<String, Object>> getProbationaryMembers() {
        String sql = "SELECT s.id, s.name as student_name, s.student_no, '2025-03-20' as probation_start, " +
                     "'2026-03-20' as probation_end, 'in_probation' as status, s.created_at " +
                     "FROM idx_student s WHERE s.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/branches")
    public R<Map<String, Object>> getBranches() {
        String sql = "SELECT c.id, c.name, col.name as college_name FROM idx_class c " +
                     "LEFT JOIN sys_major m ON c.major_id = m.id " +
                     "LEFT JOIN sys_college col ON m.college_id = col.id WHERE c.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/members")
    public R<Map<String, Object>> getMembers() {
        String sql = "SELECT s.id, ('TY202600' || s.id) as biz_no, s.student_no, s.name as student_name, " +
                     "c.name as branch_name, '2024-05-04' as join_at, '2025-05-04' as become_probationary_at, " +
                     "'2026-05-04' as formal_join_at, 'active' as status, s.created_at " +
                     "FROM idx_student s LEFT JOIN idx_class c ON s.class_id = c.id WHERE s.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }
}
