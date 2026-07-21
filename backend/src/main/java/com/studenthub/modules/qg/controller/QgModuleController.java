package com.studenthub.modules.qg.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;

@RestController
@RequestMapping("/qg")
public class QgModuleController {

    private final JdbcTemplate jdbcTemplate;

    public QgModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/difficulty-certs")
    public R<Map<String, Object>> getDifficultyCerts(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status,
            @RequestParam(required = false) String level) {

        StringBuilder where = new StringBuilder("WHERE d.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND d.cert_status = ? ");
            params.add(status);
        }
        if (level != null && !level.trim().isEmpty()) {
            where.append("AND d.level = ? ");
            params.add(level);
        }

        String countSql = "SELECT COUNT(*) FROM qg_difficulty_cert d " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT d.id, d.biz_no, d.student_id, s.name as student_name, s.student_no, " +
                "c.name as class_name, col.name as college_name, " +
                "d.level as difficulty_level, d.cert_status as status, " +
                "d.academic_year, d.proof_urls, " +
                "d.created_at, d.updated_at " +
                "FROM qg_difficulty_cert d " +
                "LEFT JOIN idx_student s ON d.student_id = s.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                where +
                "ORDER BY d.id " +
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

    @GetMapping("/difficulty-certs/{id}")
    public R<Map<String, Object>> getDifficultyCertDetail(@PathVariable Long id) {
        String sql = "SELECT d.id, d.biz_no, d.student_id, s.name as student_name, s.student_no, " +
                "c.name as class_name, col.name as college_name, " +
                "d.level as difficulty_level, d.cert_status as status, " +
                "d.academic_year, d.proof_urls, " +
                "d.created_at, d.updated_at " +
                "FROM qg_difficulty_cert d " +
                "LEFT JOIN idx_student s ON d.student_id = s.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "WHERE d.id = ? AND d.is_deleted = 0";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql, id);
        if (rows.isEmpty()) {
            return R.fail(5040, "困难认定记录不存在");
        }
        return R.ok(rows.get(0));
    }

    @GetMapping("/positions")
    public R<Map<String, Object>> getPositions(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status,
            @RequestParam(required = false) String department) {

        StringBuilder where = new StringBuilder("WHERE p.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND p.status = ? ");
            params.add(status);
        }
        if (department != null && !department.trim().isEmpty()) {
            where.append("AND p.dept_name LIKE ? ");
            params.add("%" + department + "%");
        }

        String countSql = "SELECT COUNT(*) FROM qg_position p " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT p.id, p.biz_no, p.title as position_name, " +
                "p.dept_name as department, " +
                "p.hourly_rate_cents / 100.0 as hourly_rate, " +
                "p.max_weekly_hours, p.hiring_count as quota, " +
                "p.status, p.created_at " +
                "FROM qg_position p " +
                where +
                "ORDER BY p.id " +
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

    @GetMapping("/positions/{id}")
    public R<Map<String, Object>> getPositionDetail(@PathVariable Long id) {
        String sql = "SELECT p.id, p.biz_no, p.title as position_name, " +
                "p.dept_name as department, " +
                "p.hourly_rate_cents / 100.0 as hourly_rate, " +
                "p.max_weekly_hours, p.hiring_count as quota, " +
                "p.status, p.created_at " +
                "FROM qg_position p " +
                "WHERE p.id = ? AND p.is_deleted = 0";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql, id);
        if (rows.isEmpty()) {
            return R.fail(5040, "岗位不存在");
        }
        return R.ok(rows.get(0));
    }

    @GetMapping("/statistics")
    public R<Map<String, Object>> getStatistics() {
        Map<String, Object> result = new HashMap<>();

        Integer totalPositions = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM qg_position WHERE is_deleted = 0", Integer.class);
        Integer totalCerts = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM qg_difficulty_cert WHERE is_deleted = 0", Integer.class);
        Integer activePositions = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM qg_position WHERE is_deleted = 0 AND status = 'S1'", Integer.class);

        result.put("total_positions", totalPositions != null ? totalPositions : 0);
        result.put("total_certs", totalCerts != null ? totalCerts : 0);
        result.put("active_positions", activePositions != null ? activePositions : 0);

        return R.ok(result);
    }
}
