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

        String sql = "SELECT id, student_no, name, gender, '计算机学院' as college_name, " +
                     "'软件工程' as major_name, '软工2301班' as class_name, status FROM idx_student WHERE is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total != null ? total : items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }
}
