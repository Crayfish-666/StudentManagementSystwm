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
        String sql = "SELECT RANK() OVER (ORDER BY c.total_score DESC) as rank, " +
                     "s.name as student_name, s.student_no, IFNULL(col.name, '计算机学院') as college_name, " +
                     "c.total_score, c.academic_score, c.ty_score as moral_score, c.st_score as sports_score, " +
                     "c.sq_score as art_score, c.qg_score as labor_score " +
                     "FROM cmp_score c " +
                     "LEFT JOIN idx_student s ON c.student_id = s.id " +
                     "LEFT JOIN sys_college col ON s.college_id = col.id " +
                     "ORDER BY c.total_score DESC";
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
        String sql = "SELECT c.id, s.name as student_name, c.academic_score, c.ty_score as moral_score, " +
                     "c.st_score as physical_score, c.st_score as sports_score, c.sq_score as art_score, " +
                     "c.qg_score as labor_score, c.total_score " +
                     "FROM cmp_score c LEFT JOIN idx_student s ON c.student_id = s.id";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }
}
