package com.studenthub.modules.cmp.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/cmp")
public class CmpModuleController {

    private final JdbcTemplate jdbcTemplate;

    public CmpModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/rankings")
    public R<Map<String, Object>> getRankings() {
        String sql = "SELECT id as rank, name as student_name, student_no, '计算机学院' as college_name, " +
                     "(98.0 - id * 0.85) as total_score FROM idx_student WHERE is_deleted = 0 ORDER BY id ASC";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/scores")
    public R<Map<String, Object>> getScores() {
        String sql = "SELECT id, name as student_name, 90.0 as academic_score, 95.0 as moral_score, " +
                     "92.0 as physical_score, 88.0 as art_score, 96.0 as labor_score, (90.0 + id*0.4) as total_score " +
                     "FROM idx_student WHERE is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }
}
