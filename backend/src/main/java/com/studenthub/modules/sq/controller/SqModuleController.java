package com.studenthub.modules.sq.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;

@RestController
@RequestMapping("/sq")
public class SqModuleController {

    private final JdbcTemplate jdbcTemplate;

    public SqModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/incidents")
    public R<Map<String, Object>> getIncidents(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status,
            @RequestParam(required = false) String level,
            @RequestParam(required = false) Long building_id) {

        StringBuilder where = new StringBuilder("WHERE i.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND i.status = ? ");
            params.add(status);
        }
        if (level != null && !level.trim().isEmpty()) {
            where.append("AND i.level = ? ");
            params.add(level);
        }
        if (building_id != null) {
            where.append("AND i.building_id = ? ");
            params.add(building_id);
        }

        String countSql = "SELECT COUNT(*) FROM sq_incident i " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT i.id, i.biz_no, i.incident_type as type, i.level, i.status, " +
                "i.building_id, b.name as building_name, " +
                "i.reporter_id, i.handler_id, " +
                "i.description, i.resolution, i.closed_at, " +
                "i.created_at, i.updated_at " +
                "FROM sq_incident i " +
                "LEFT JOIN sq_building b ON i.building_id = b.id " +
                where +
                "ORDER BY i.id " +
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

    @GetMapping("/incidents/{id}")
    public R<Map<String, Object>> getIncidentDetail(@PathVariable Long id) {
        String sql = "SELECT i.id, i.biz_no, i.incident_type as type, i.level, i.status, " +
                "i.building_id, b.name as building_name, " +
                "i.reporter_id, i.handler_id, " +
                "i.description, i.resolution, i.closed_at, " +
                "i.created_at, i.updated_at " +
                "FROM sq_incident i " +
                "LEFT JOIN sq_building b ON i.building_id = b.id " +
                "WHERE i.id = ? AND i.is_deleted = 0";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql, id);
        if (rows.isEmpty()) {
            return R.fail(4040, "事件不存在");
        }
        return R.ok(rows.get(0));
    }

    @GetMapping("/inspections")
    public R<Map<String, Object>> getInspections(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String status,
            @RequestParam(required = false) Long building_id) {

        StringBuilder where = new StringBuilder("WHERE i.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (status != null && !status.trim().isEmpty()) {
            where.append("AND i.status = ? ");
            params.add(status);
        }
        if (building_id != null) {
            where.append("AND i.building_id = ? ");
            params.add(building_id);
        }

        String countSql = "SELECT COUNT(*) FROM sq_incident i " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT i.id, i.biz_no, i.incident_type as type, i.level, i.status, " +
                "i.building_id, b.name as building_name, " +
                "i.reporter_id as inspector_id, " +
                "i.description, i.resolution, i.closed_at as resolved_at, " +
                "i.created_at as patrol_time, i.updated_at " +
                "FROM sq_incident i " +
                "LEFT JOIN sq_building b ON i.building_id = b.id " +
                where +
                "ORDER BY i.id " +
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

    @GetMapping("/buildings")
    public R<Map<String, Object>> getBuildings(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {

        String countSql = "SELECT COUNT(*) FROM sq_building WHERE is_deleted = 0";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT id, code, name, total_floors, tutor_id, created_at, is_deleted " +
                "FROM sq_building WHERE is_deleted = 0 " +
                "ORDER BY id " +
                "LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/buildings/tree")
    public R<List<Map<String, Object>>> getBuildingTree() {
        String buildingSql = "SELECT id, code, name, total_floors FROM sq_building WHERE is_deleted = 0 ORDER BY id";
        List<Map<String, Object>> buildings = jdbcTemplate.queryForList(buildingSql);

        List<Map<String, Object>> tree = new ArrayList<>();
        for (Map<String, Object> building : buildings) {
            Integer buildingId = (Integer) building.get("id");
            Integer totalFloors = (Integer) building.get("total_floors");
            if (totalFloors == null) totalFloors = 6;

            List<Map<String, Object>> floors = new ArrayList<>();
            for (int i = 1; i <= totalFloors; i++) {
                Map<String, Object> floor = new HashMap<>();
                floor.put("id", buildingId * 100 + i);
                floor.put("name", i + "楼");
                floor.put("floor_no", i);
                floors.add(floor);
            }

            Map<String, Object> node = new HashMap<>(building);
            node.put("children", floors);
            tree.add(node);
        }
        return R.ok(tree);
    }

    @GetMapping("/statistics")
    public R<Map<String, Object>> getStatistics() {
        Map<String, Object> result = new HashMap<>();

        Integer totalBuildings = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM sq_building WHERE is_deleted = 0", Integer.class);
        Integer totalIncidents = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM sq_incident WHERE is_deleted = 0", Integer.class);
        Integer openIncidents = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM sq_incident WHERE is_deleted = 0 AND status != 'closed'", Integer.class);

        result.put("total_buildings", totalBuildings != null ? totalBuildings : 0);
        result.put("total_incidents", totalIncidents != null ? totalIncidents : 0);
        result.put("open_incidents", openIncidents != null ? openIncidents : 0);

        return R.ok(result);
    }
}
