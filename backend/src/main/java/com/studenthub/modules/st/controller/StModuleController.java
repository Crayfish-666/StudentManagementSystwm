package com.studenthub.modules.st.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;
import java.time.LocalDateTime;

@RestController
@RequestMapping("/st")
public class StModuleController {

    private final JdbcTemplate jdbcTemplate;

    public StModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/activities")
    public R<Map<String, Object>> getActivities(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status,
            @RequestParam(required = false) String category) {

        StringBuilder where = new StringBuilder("WHERE a.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND a.activity_status = ? ");
            params.add(status);
        }

        String countSql = "SELECT COUNT(*) FROM st_activity a " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT a.id, a.biz_no, a.title as name, a.level as category, a.activity_status as status, " +
                "a.start_time, a.end_time, a.location, " +
                "a.budget_cents, " +
                "s.name as association_name, s.id as association_id, " +
                "a.created_at, a.updated_at " +
                "FROM st_activity a " +
                "LEFT JOIN st_association s ON a.assoc_id = s.id " +
                where +
                "ORDER BY a.id " +
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

    @GetMapping("/activities/{id}")
    public R<Map<String, Object>> getActivityDetail(@PathVariable Long id) {
        String sql = "SELECT a.id, a.biz_no, a.title as name, a.level as category, a.activity_status as status, " +
                "a.start_time, a.end_time, a.location, " +
                "a.budget_cents, a.emergency_plan_url, a.safety_commitment_url, " +
                "s.name as association_name, s.id as association_id, " +
                "a.created_at, a.updated_at " +
                "FROM st_activity a " +
                "LEFT JOIN st_association s ON a.assoc_id = s.id " +
                "WHERE a.id = ? AND a.is_deleted = 0";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql, id);
        if (rows.isEmpty()) {
            return R.fail(3040, "活动不存在");
        }
        return R.ok(rows.get(0));
    }

    @GetMapping("/associations")
    public R<Map<String, Object>> getAssociations(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status,
            @RequestParam(required = false) String type) {

        StringBuilder where = new StringBuilder("WHERE s.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND s.status = ? ");
            params.add(status);
        }

        String countSql = "SELECT COUNT(*) FROM st_association s " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT s.id, s.assoc_code as biz_no, s.name, s.assoc_code as code, " +
                "s.status, s.star_rating, s.president_id, s.college_id, " +
                "s.created_at as founded_at, " +
                "s.created_at, s.updated_at " +
                "FROM st_association s " +
                where +
                "ORDER BY s.id " +
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

    @GetMapping("/associations/{id}")
    public R<Map<String, Object>> getAssociationDetail(@PathVariable Long id) {
        String sql = "SELECT s.id, s.assoc_code as biz_no, s.name, s.assoc_code as code, " +
                "s.status, s.star_rating, s.president_id, s.college_id, " +
                "s.created_at as founded_at, " +
                "s.created_at, s.updated_at " +
                "FROM st_association s " +
                "WHERE s.id = ? AND s.is_deleted = 0";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql, id);
        if (rows.isEmpty()) {
            return R.fail(3040, "社团不存在");
        }
        return R.ok(rows.get(0));
    }

    @GetMapping("/recruit-plans")
    public R<Map<String, Object>> getRecruitPlans(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status,
            @RequestParam(required = false) Long association_id) {

        StringBuilder where = new StringBuilder("WHERE r.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND r.status = ? ");
            params.add(status);
        }
        if (association_id != null) {
            where.append("AND r.assoc_id = ? ");
            params.add(association_id);
        }

        String countSql = "SELECT COUNT(*) FROM st_recruit_plan r " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT r.id, r.title, r.status, " +
                "r.target_count as recruit_count, r.accepted_count, " +
                "s.name as association_name, s.id as association_id, " +
                "r.created_at, r.updated_at " +
                "FROM st_recruit_plan r " +
                "LEFT JOIN st_association s ON r.assoc_id = s.id " +
                where +
                "ORDER BY r.id " +
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

    @GetMapping("/recruit-plans/{id}")
    public R<Map<String, Object>> getRecruitPlanDetail(@PathVariable Long id) {
        String sql = "SELECT r.id, r.title, r.status, " +
                "r.target_count as recruit_count, r.accepted_count, " +
                "r.is_finished, r.finished_reason, " +
                "s.name as association_name, s.id as association_id, " +
                "r.created_at, r.updated_at " +
                "FROM st_recruit_plan r " +
                "LEFT JOIN st_association s ON r.assoc_id = s.id " +
                "WHERE r.id = ? AND r.is_deleted = 0";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql, id);
        if (rows.isEmpty()) {
            return R.fail(3040, "招新计划不存在");
        }
        return R.ok(rows.get(0));
    }

    // 招新申请/招新广场列表
    @GetMapping("/recruit-applies")
    public R<Map<String, Object>> getRecruitApplies(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status) {

        StringBuilder where = new StringBuilder("WHERE ra.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND ra.status = ? ");
            params.add(status);
        }

        String countSql = "SELECT COUNT(*) FROM st_recruit_apply ra " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT ra.id, ra.plan_id, p.title as plan_title, " +
                "ra.assoc_id, s.name as association_name, " +
                "ra.student_id, stu.name as student_name, stu.student_no, " +
                "col.name as college_name, " +
                "ra.apply_reason, ra.status, ra.interview_score, ra.created_at " +
                "FROM st_recruit_apply ra " +
                "LEFT JOIN st_recruit_plan p ON ra.plan_id = p.id " +
                "LEFT JOIN st_association s ON ra.assoc_id = s.id " +
                "LEFT JOIN idx_student stu ON ra.student_id = stu.id " +
                "LEFT JOIN sys_college col ON stu.college_id = col.id " +
                where +
                "ORDER BY ra.id DESC " +
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

    @GetMapping("/statistics")
    public R<Map<String, Object>> getStatistics() {
        Map<String, Object> result = new HashMap<>();

        Integer totalActivities = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM st_activity WHERE is_deleted = 0", Integer.class);
        Integer totalAssociations = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM st_association WHERE is_deleted = 0", Integer.class);
        Integer totalRecruits = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM st_recruit_plan WHERE is_deleted = 0", Integer.class);

        result.put("total_activities", totalActivities != null ? totalActivities : 0);
        result.put("total_associations", totalAssociations != null ? totalAssociations : 0);
        result.put("total_recruits", totalRecruits != null ? totalRecruits : 0);

        return R.ok(result);
    }
    @PostMapping("/associations")
    public R<Void> createAssociation(@RequestBody Map<String, Object> body) {
        String sql = "INSERT INTO st_association (name, assoc_code, status, star_rating, president_id, college_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)";
        jdbcTemplate.update(sql, body.get("name"), body.get("assoc_code"), body.getOrDefault("status", "active"), body.get("star_rating"), body.get("president_id"), body.get("college_id"), LocalDateTime.now(), LocalDateTime.now());
        return R.ok();
    }

    @PutMapping("/associations/{id}")
    public R<Void> updateAssociation(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String sql = "UPDATE st_association SET name = ?, status = ?, star_rating = ?, president_id = ?, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, body.get("name"), body.get("status"), body.get("star_rating"), body.get("president_id"), LocalDateTime.now(), id);
        return R.ok();
    }

    @DeleteMapping("/associations/{id}")
    public R<Void> deleteAssociation(@PathVariable Long id) {
        String sql = "UPDATE st_association SET is_deleted = 1, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/activities")
    public R<Void> createActivity(@RequestBody Map<String, Object> body) {
        String sql = "INSERT INTO st_activity (biz_no, title, level, activity_status, start_time, end_time, location, budget_cents, assoc_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)";
        String bizNo = "ACT" + System.currentTimeMillis();
        jdbcTemplate.update(sql, bizNo, body.get("title"), body.get("level"), body.getOrDefault("activity_status", "S0"), body.get("start_time"), body.get("end_time"), body.get("location"), body.get("budget_cents"), body.get("assoc_id"), LocalDateTime.now(), LocalDateTime.now());
        return R.ok();
    }

    @PutMapping("/activities/{id}")
    public R<Void> updateActivity(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String sql = "UPDATE st_activity SET title = ?, level = ?, start_time = ?, end_time = ?, location = ?, budget_cents = ?, updated_at = ? WHERE id = ? AND activity_status = 'S0' AND is_deleted = 0";
        jdbcTemplate.update(sql, body.get("title"), body.get("level"), body.get("start_time"), body.get("end_time"), body.get("location"), body.get("budget_cents"), LocalDateTime.now(), id);
        return R.ok();
    }

    @DeleteMapping("/activities/{id}")
    public R<Void> deleteActivity(@PathVariable Long id) {
        String sql = "UPDATE st_activity SET is_deleted = 1, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/activities/{id}/submit")
    public R<Void> submitActivity(@PathVariable Long id) {
        String sql = "UPDATE st_activity SET activity_status = 'S1', updated_at = ? WHERE id = ? AND activity_status = 'S0' AND is_deleted = 0";
        jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/activities/{id}/approve")
    public R<Void> approveActivity(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String action = (String) body.get("action");
        String newStatus = "approve".equals(action) ? "S3" : "S4";
        String sql = "UPDATE st_activity SET activity_status = ?, updated_at = ? WHERE id = ? AND activity_status = 'S1' AND is_deleted = 0";
        jdbcTemplate.update(sql, newStatus, LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/activities/{id}/withdraw")
    public R<Void> withdrawActivity(@PathVariable Long id) {
        String sql = "UPDATE st_activity SET activity_status = 'S0', updated_at = ? WHERE id = ? AND activity_status = 'S1' AND is_deleted = 0";
        jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/recruit-plans")
    public R<Void> createRecruitPlan(@RequestBody Map<String, Object> body) {
        String sql = "INSERT INTO st_recruit_plan (title, assoc_id, target_count, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)";
        jdbcTemplate.update(sql, body.get("title"), body.get("assoc_id"), body.get("target_count"), body.getOrDefault("status", "S0"), LocalDateTime.now(), LocalDateTime.now());
        return R.ok();
    }

    @PutMapping("/recruit-plans/{id}")
    public R<Void> updateRecruitPlan(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String sql = "UPDATE st_recruit_plan SET title = ?, target_count = ?, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, body.get("title"), body.get("target_count"), LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/recruit-plans/{id}/submit")
    public R<Void> submitRecruitPlan(@PathVariable Long id) {
        String sql = "UPDATE st_recruit_plan SET status = 'S1', updated_at = ? WHERE id = ? AND status = 'S0' AND is_deleted = 0";
        jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/recruit-plans/{id}/approve")
    public R<Void> approveRecruitPlan(@PathVariable Long id) {
        String sql = "UPDATE st_recruit_plan SET status = 'S3', updated_at = ? WHERE id = ? AND status = 'S1' AND is_deleted = 0";
        jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/recruit-plans/{id}/reject")
    public R<Void> rejectRecruitPlan(@PathVariable Long id) {
        String sql = "UPDATE st_recruit_plan SET status = 'S4', updated_at = ? WHERE id = ? AND status = 'S1' AND is_deleted = 0";
        jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/recruit-applies")
    public R<Void> createRecruitApply(@RequestBody Map<String, Object> body) {
        String sql = "INSERT INTO st_recruit_apply (plan_id, assoc_id, student_id, apply_reason, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)";
        jdbcTemplate.update(sql, body.get("plan_id"), body.get("assoc_id"), body.get("student_id"), body.get("apply_reason"), body.getOrDefault("status", "pending"), LocalDateTime.now(), LocalDateTime.now());
        return R.ok();
    }
}

