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
        String sql = "SELECT s.id, s.name as student_name, s.student_no, '特别困难' as level, " +
                     "'approved' as status FROM idx_student s WHERE s.is_deleted = 0";
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
        String countSql = "SELECT COUNT(*) FROM qg_position WHERE is_deleted = 0";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);

        String sql = "SELECT id, biz_no, dept_name, title, (hourly_rate_cents / 100.0) as hourly_rate, " +
                     "hiring_count as quota, status FROM qg_position WHERE is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total != null ? total : items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/attendances")
    public R<Map<String, Object>> getAttendances() {
        String sql = "SELECT s.id, s.name as student_name, '图书整理助理' as position_title, " +
                     "'2026-03-18' as work_date, 4.0 as hours, 'approved' as status FROM idx_student s WHERE s.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }
}
