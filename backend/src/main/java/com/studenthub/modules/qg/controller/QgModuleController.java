package com.studenthub.modules.qg.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
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

    // 工时打卡列表
    @GetMapping("/attendances")
    public R<Map<String, Object>> getAttendances(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status) {

        StringBuilder where = new StringBuilder("WHERE a.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND a.status = ? ");
            params.add(status);
        }

        String countSql = "SELECT COUNT(*) FROM qg_attendance a " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT a.id, a.position_id, p.title as position_name, p.dept_name as department, " +
                "a.student_id, s.name as student_name, s.student_no, " +
                "a.clock_in, a.clock_out, a.hours, a.status, a.work_content, a.created_at " +
                "FROM qg_attendance a " +
                "LEFT JOIN qg_position p ON a.position_id = p.id " +
                "LEFT JOIN idx_student s ON a.student_id = s.id " +
                where +
                "ORDER BY a.id DESC " +
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

    // 岗位申请列表
    @GetMapping("/applies")
    public R<Map<String, Object>> getApplies(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status) {

        StringBuilder where = new StringBuilder("WHERE pa.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND pa.status = ? ");
            params.add(status);
        }

        String countSql = "SELECT COUNT(*) FROM qg_position_apply pa " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT pa.id, pa.position_id, p.title as position_name, p.dept_name as department, " +
                "pa.student_id, s.name as student_name, s.student_no, " +
                "pa.status, pa.apply_reason, pa.created_at " +
                "FROM qg_position_apply pa " +
                "LEFT JOIN qg_position p ON pa.position_id = p.id " +
                "LEFT JOIN idx_student s ON pa.student_id = s.id " +
                where +
                "ORDER BY pa.id DESC " +
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
    @PostMapping("/positions")
    public R<Map<String, Object>> createPosition(@RequestBody Map<String, Object> body) {
        String bizNo = "QGP" + System.currentTimeMillis();
        String sql = "INSERT INTO qg_position (biz_no, title, dept_name, hourly_rate_cents, max_weekly_hours, hiring_count, status, created_at, updated_at, is_deleted) VALUES (?, ?, ?, ?, ?, ?, 'S1', ?, ?, 0)";
        LocalDateTime now = LocalDateTime.now();
        jdbcTemplate.update(sql, bizNo, body.get("title"), body.get("dept_name"), body.get("hourly_rate_cents"), body.get("max_weekly_hours"), body.get("hiring_count"), now, now);
        body.put("biz_no", bizNo);
        body.put("status", "S1");
        return R.ok(body);
    }

    @PutMapping("/positions/{id}")
    public R<Void> updatePosition(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String sql = "UPDATE qg_position SET title = ?, dept_name = ?, hourly_rate_cents = ?, max_weekly_hours = ?, hiring_count = ?, status = ?, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, body.get("title"), body.get("dept_name"), body.get("hourly_rate_cents"), body.get("max_weekly_hours"), body.get("hiring_count"), body.get("status"), LocalDateTime.now(), id);
        return R.ok();
    }

    @DeleteMapping("/positions/{id}")
    public R<Void> deletePosition(@PathVariable Long id) {
        String sql = "UPDATE qg_position SET is_deleted = 1, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/difficulty-certs")
    public R<Map<String, Object>> createDifficultyCert(@RequestBody Map<String, Object> body) {
        String bizNo = "QGC" + System.currentTimeMillis();
        String sql = "INSERT INTO qg_difficulty_cert (biz_no, student_id, level, academic_year, cert_status, proof_urls, created_at, updated_at, is_deleted) VALUES (?, ?, ?, ?, 'pending', ?, ?, ?, 0)";
        LocalDateTime now = LocalDateTime.now();
        jdbcTemplate.update(sql, bizNo, body.get("student_id"), body.get("level"), body.get("academic_year"), body.get("proof_urls"), now, now);
        body.put("biz_no", bizNo);
        body.put("cert_status", "pending");
        return R.ok(body);
    }

    @PutMapping("/difficulty-certs/{id}")
    public R<Void> updateDifficultyCert(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String sql = "UPDATE qg_difficulty_cert SET cert_status = ?, level = ?, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, body.get("cert_status"), body.get("level"), LocalDateTime.now(), id);
        return R.ok();
    }

    @DeleteMapping("/difficulty-certs/{id}")
    public R<Void> deleteDifficultyCert(@PathVariable Long id) {
        String sql = "UPDATE qg_difficulty_cert SET is_deleted = 1, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/attendances")
    public R<Map<String, Object>> createAttendance(@RequestBody Map<String, Object> body) {
        String sql = "INSERT INTO qg_attendance (position_id, student_id, clock_in, clock_out, hours, status, work_content, created_at, updated_at, is_deleted) VALUES (?, ?, ?, ?, ?, 'pending', ?, ?, ?, 0)";
        LocalDateTime now = LocalDateTime.now();
        jdbcTemplate.update(sql, body.get("position_id"), body.get("student_id"), body.get("clock_in"), body.get("clock_out"), body.get("hours"), body.get("work_content"), now, now);
        body.put("status", "pending");
        return R.ok(body);
    }

    @PostMapping("/applies")
    public R<Map<String, Object>> createApply(@RequestBody Map<String, Object> body) {
        String sql = "INSERT INTO qg_position_apply (position_id, student_id, apply_reason, status, created_at, updated_at, is_deleted) VALUES (?, ?, ?, 'pending', ?, ?, 0)";
        LocalDateTime now = LocalDateTime.now();
        jdbcTemplate.update(sql, body.get("position_id"), body.get("student_id"), body.get("apply_reason"), now, now);
        body.put("status", "pending");
        return R.ok(body);
    }

    @PutMapping("/applies/{id}")
    public R<Void> updateApplyStatus(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String sql = "UPDATE qg_position_apply SET status = ?, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, body.get("status"), LocalDateTime.now(), id);
        return R.ok();
    }
}
