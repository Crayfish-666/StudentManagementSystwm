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
}
