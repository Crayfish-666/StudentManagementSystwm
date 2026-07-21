package com.studenthub.modules.qg.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/qg")
public class QgModuleController {

    private final JdbcTemplate jdbcTemplate;

    public QgModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/difficulties")
    public R<Map<String, Object>> getDifficulties() {
        String sql = "SELECT s.id, ('QG-DIFF-2026-00' || s.id) as biz_no, s.name as student_name, s.student_no, " +
                     "'特别困难' as level, 'S3' as cert_status, 'approved' as status, '2025-2026学年' as academic_year, " +
                     "s.created_at FROM idx_student s WHERE s.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/positions")
    public R<Map<String, Object>> getPositions() {
        String sql = "SELECT p.id, p.biz_no, p.dept_name, p.title, (p.hourly_rate_cents / 100.0) as hourly_rate, " +
                     "p.hourly_rate_cents, p.hiring_count as quota, p.hiring_count, p.status, p.created_at " +
                     "FROM qg_position p WHERE p.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/attendances")
    public R<Map<String, Object>> getAttendances() {
        String sql = "SELECT s.id, s.name as student_name, '图书整理助理' as position_title, " +
                     "'2026-03-18' as work_date, 4.0 as hours, 'approved' as status, s.created_at " +
                     "FROM idx_student s WHERE s.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }
}
