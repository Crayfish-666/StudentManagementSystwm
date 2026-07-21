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

        String sql = "SELECT t.id, t.biz_no, s.name as student_name, s.student_no, '计算机2301团支部' as branch_name, " +
                     "'计算机学院' as college_name, t.apply_date, t.app_status as status " +
                     "FROM ty_application t LEFT JOIN idx_student s ON t.student_id = s.id " +
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
        String sql = "SELECT t.id, t.biz_no, s.name as student_name, s.student_no, '计算机2301团支部' as branch_name, " +
                     "'计算机学院' as college_name, t.apply_date, t.app_status as status " +
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
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> m = new HashMap<>();
            m.put("id", (long) i);
            m.put("title", String.format("2026春季支部推优大会第%d期", i));
            m.put("branch_name", "计算机2301团支部");
            m.put("meeting_date", String.format("2026-03-%02d", (i % 28) + 1));
            m.put("attendee_rate", String.format("%.1f%%", 90.0 + (i % 10)));
            m.put("status", i % 2 == 0 ? "completed" : "scheduled");
            items.add(m);
        }
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", 20);
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/cultivation-links")
    public R<Map<String, Object>> getCultivationLinks() {
        String sql = "SELECT id, name as student_name, student_no, '张辅导员' as cultivator_name, " +
                     "'培养联系期' as stage, '85%' as progress FROM idx_student WHERE is_deleted = 0";
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
        String sql = "SELECT id, name as student_name, student_no, '计算机学院' as college_name, " +
                     "'2026-03-01' as approval_date, 'active' as status FROM idx_student WHERE is_deleted = 0";
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
        String sql = "SELECT id, name as student_name, student_no, '合格' as result, " +
                     "'张支部书记' as reviewer, '2026-03-15' as review_date FROM idx_student WHERE is_deleted = 0";
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
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> m = new HashMap<>();
            m.put("id", (long) i);
            m.put("meeting_title", String.format("2026上半年预备团员接收大会第%d期", i));
            m.put("meeting_date", String.format("2026-03-%02d", (i % 25) + 1));
            m.put("pass_count", 10 + i);
            items.add(m);
        }
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", 20);
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/probationary-members")
    public R<Map<String, Object>> getProbationaryMembers() {
        String sql = "SELECT id, name as student_name, student_no, '2025-03-20' as probation_start, " +
                     "'2026-03-20' as probation_end, 'ready_for_regular' as status FROM idx_student WHERE is_deleted = 0";
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
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> b = new HashMap<>();
            b.put("id", (long) i);
            b.put("name", String.format("2023级团支部第%d分部", i));
            items.add(b);
        }
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", 20);
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/members")
    public R<Map<String, Object>> getMembers() {
        String sql = "SELECT id, name as student_name, student_no, 'TY2026' || id as league_no, " +
                     "'2024-05-04' as join_date FROM idx_student WHERE is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }
}
