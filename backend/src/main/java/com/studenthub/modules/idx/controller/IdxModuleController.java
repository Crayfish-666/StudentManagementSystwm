package com.studenthub.modules.idx.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/idx")
public class IdxModuleController {

    private final JdbcTemplate jdbcTemplate;

    public IdxModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/students")
    public R<Map<String, Object>> getStudents() {
        String countSql = "SELECT COUNT(*) FROM idx_student WHERE is_deleted = 0";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);

        String sql = "SELECT s.id, s.student_no, s.name, s.name as student_name, " +
                     "CASE WHEN s.gender = 'M' THEN '男' ELSE '女' END as gender, " +
                     "col.name as college_name, m.name as major_name, c.name as class_name, " +
                     "s.status, s.created_at " +
                     "FROM idx_student s " +
                     "LEFT JOIN sys_college col ON s.college_id = col.id " +
                     "LEFT JOIN sys_major m ON s.major_id = m.id " +
                     "LEFT JOIN idx_class c ON s.class_id = c.id " +
                     "WHERE s.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total != null ? total : items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }
}
