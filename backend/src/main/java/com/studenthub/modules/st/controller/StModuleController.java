package com.studenthub.modules.st.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;

@RestController
@RequestMapping("/st")
public class StModuleController {

    private final JdbcTemplate jdbcTemplate;

    public StModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/associations")
    public R<Map<String, Object>> getAssociations() {
        String sql = "SELECT a.id, a.assoc_code, a.name, '学术科技类' as category, s.name as president_name, " +
                     "'王辅导员' as tutor_name, a.star_rating as star_level, a.star_rating, a.status, " +
                     "50 as member_count, a.created_at FROM st_association a " +
                     "LEFT JOIN idx_student s ON a.president_id = s.id WHERE a.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/users")
    public R<List<Map<String, Object>>> getUsers() {
        String sql = "SELECT id, username, display_name FROM sys_user WHERE user_type != 'student' AND is_deleted = 0";
        List<Map<String, Object>> users = jdbcTemplate.queryForList(sql);
        return R.ok(users);
    }

    @GetMapping("/students")
    public R<List<Map<String, Object>>> getStudents() {
        String sql = "SELECT id, student_no, name FROM idx_student WHERE is_deleted = 0";
        List<Map<String, Object>> list = jdbcTemplate.queryForList(sql);
        return R.ok(list);
    }

    @GetMapping("/associations/{id}/founders")
    public R<List<Map<String, Object>>> getFounders(@PathVariable Long id) {
        String sql = "SELECT s.id, s.name as student_name, s.student_no, '发起人' as role " +
                     "FROM idx_student s WHERE s.is_deleted = 0 LIMIT 3";
        List<Map<String, Object>> founders = jdbcTemplate.queryForList(sql);
        return R.ok(founders);
    }

    @GetMapping("/associations/{id}/members")
    public R<List<Map<String, Object>>> getMembers(@PathVariable Long id) {
        String sql = "SELECT s.id, s.name as student_name, s.student_no, " +
                     "CASE WHEN s.id = 1 THEN '社长' ELSE '社员' END as role " +
                     "FROM idx_student s WHERE s.is_deleted = 0 LIMIT 10";
        List<Map<String, Object>> members = jdbcTemplate.queryForList(sql);
        return R.ok(members);
    }

    @GetMapping("/recruit-plans")
    public R<Map<String, Object>> getRecruitPlans() {
        String sql = "SELECT a.id, ('2026春季' || a.name || '招新计划') as title, a.name as association_name, " +
                     "50 as target_count, 32 as applied_count, 20 as accepted_count, 'S3' as status, a.created_at " +
                     "FROM st_association a WHERE a.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/recruit-applies")
    public R<Map<String, Object>> getRecruitApplies() {
        String sql = "SELECT s.id, s.name as student_name, s.student_no, '计算机算法与编程社' as association_name, " +
                     "'approved' as status, s.created_at FROM idx_student s WHERE s.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/activities")
    public R<Map<String, Object>> getActivities() {
        String sql = "SELECT a.id, ('ST-ACT-2026-00' || a.id) as biz_no, ('2026' || a.name || '主题活动') as title, " +
                     "a.name as association_name, '2026-04-15' as activity_date, '学生活动中心401室' as location, " +
                     "'S3' as status, 'B' as level, 50000 as budget_cents, a.created_at " +
                     "FROM st_association a WHERE a.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }
}
