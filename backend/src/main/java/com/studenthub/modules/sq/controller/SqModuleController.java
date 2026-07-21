package com.studenthub.modules.sq.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/sq")
public class SqModuleController {

    private final JdbcTemplate jdbcTemplate;

    public SqModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/buildings")
    public R<Map<String, Object>> getBuildings() {
        String sql = "SELECT b.id, b.code, b.name, b.total_floors, (100 + b.id*2) as room_count, " +
                     "'张宿管辅导员' as tutor_name, b.created_at FROM sq_building b WHERE b.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/inspections")
    public R<Map<String, Object>> getInspections() {
        String sql = "SELECT b.id, b.name as building_name, '302' as room_no, 95 as score, " +
                     "'王网格员' as inspector_name, '2026-03-18' as inspect_date, b.created_at " +
                     "FROM sq_building b WHERE b.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/incidents")
    public R<Map<String, Object>> getIncidents() {
        String sql = "SELECT b.id, ('SQ-INC-2026-00' || b.id) as biz_no, b.name as building_name, '405' as room_no, " +
                     "'大功率违规用电' as type, '大功率违规用电' as incident_type, 'resolved' as status, " +
                     "'宿管巡查发现大功率违规用电，已登记处理。' as description, b.created_at " +
                     "FROM sq_building b WHERE b.is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);
        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }
}
