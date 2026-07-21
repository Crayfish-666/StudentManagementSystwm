package com.studenthub.modules.st.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/st")
public class StModuleController {

    private final JdbcTemplate jdbcTemplate;

    public StModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/associations")
    public R<Map<String, Object>> getAssociations() {
        String countSql = "SELECT COUNT(*) FROM st_association WHERE is_deleted = 0";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);

        String sql = "SELECT a.id, a.assoc_code, a.name, '学术科技类' as category, s.name as president_name, " +
                     "a.star_rating as star_level, a.status FROM st_association a " +
                     "LEFT JOIN idx_student s ON a.president_id = s.id WHERE a.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total != null ? total : items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/recruit-plans")
    public R<Map<String, Object>> getRecruitPlans() {
        String sql = "SELECT a.id, '2026春季招新计划' as title, a.name as association_name, 50 as target_count, " +
                     "32 as applied_count, 'recruiting' as status FROM st_association a WHERE a.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/recruit-applies")
    public R<Map<String, Object>> getRecruitApplies() {
        String sql = "SELECT s.id, s.name as student_name, s.student_no, '计算机算法与编程社' as association_name, " +
                     "'pending' as status FROM idx_student s WHERE s.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/activities")
    public R<Map<String, Object>> getActivities() {
        String sql = "SELECT a.id, '第十二届全校社团主题活动' as title, a.name as association_name, '2026-04-15' as activity_date, " +
                     "'学生活动中心401' as location, 'approved' as status FROM st_association a WHERE a.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }
}
