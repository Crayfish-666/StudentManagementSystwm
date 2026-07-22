package com.studenthub.modules.sq.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
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

        StringBuilder where = new StringBuilder("WHERE 1=1 ");
        List<Object> params = new ArrayList<>();

        if (building_id != null) {
            where.append("AND i.building_id = ? ");
            params.add(building_id);
        }

        String countSql = "SELECT COUNT(*) FROM sq_inspection i " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT i.id, i.biz_no, i.building_id, b.name as building_name, " +
                "i.room_no, i.score, i.hygiene_status, i.safety_status, " +
                "i.inspector_name, i.remark, i.patrol_time, i.created_at " +
                "FROM sq_inspection i " +
                "LEFT JOIN sq_building b ON i.building_id = b.id " +
                where +
                "ORDER BY i.id DESC " +
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

        String roomSql = "SELECT id, building_id, room_no, floor_no, bed_count, occupied_count, status FROM sq_room WHERE is_deleted = 0";
        List<Map<String, Object>> rooms = jdbcTemplate.queryForList(roomSql);

        List<Map<String, Object>> tree = new ArrayList<>();
        for (Map<String, Object> building : buildings) {
            Integer buildingId = (Integer) building.get("id");
            Integer totalFloors = (Integer) building.get("total_floors");
            if (totalFloors == null) totalFloors = 6;

            List<Map<String, Object>> floors = new ArrayList<>();
            for (int f = 1; f <= totalFloors; f++) {
                Map<String, Object> floor = new HashMap<>();
                floor.put("id", buildingId * 100 + f);
                floor.put("name", f + "楼");
                floor.put("floor_no", f);

                List<Map<String, Object>> floorRooms = new ArrayList<>();
                for (Map<String, Object> r : rooms) {
                    Integer bId = (Integer) r.get("building_id");
                    Integer fNo = (Integer) r.get("floor_no");
                    if (bId != null && bId.equals(buildingId) && fNo != null && fNo.equals(f)) {
                        floorRooms.add(r);
                    }
                }
                floor.put("children", floorRooms);
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
    @PostMapping("/incidents")
    public R<Map<String, Object>> createIncident(@RequestBody Map<String, Object> body) {
        String bizNo = "SQI" + System.currentTimeMillis();
        String sql = "INSERT INTO sq_incident (biz_no, incident_type, level, building_id, reporter_id, description, status, created_at, updated_at, is_deleted) VALUES (?, ?, ?, ?, ?, ?, 'open', ?, ?, 0)";
        LocalDateTime now = LocalDateTime.now();
        jdbcTemplate.update(sql, bizNo, body.get("incident_type"), body.get("level"), body.get("building_id"), body.get("reporter_id"), body.get("description"), now, now);
        body.put("biz_no", bizNo);
        body.put("status", "open");
        return R.ok(body);
    }

    @PutMapping("/incidents/{id}")
    public R<Void> updateIncident(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String sql = "UPDATE sq_incident SET status = ?, handler_id = ?, resolution = ?, closed_at = ?, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, body.get("status"), body.get("handler_id"), body.get("resolution"), body.get("closed_at"), LocalDateTime.now(), id);
        return R.ok();
    }

    @DeleteMapping("/incidents/{id}")
    public R<Void> deleteIncident(@PathVariable Long id) {
        String sql = "UPDATE sq_incident SET is_deleted = 1, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return R.ok();
    }

    @PostMapping("/inspections")
    public R<Map<String, Object>> createInspection(@RequestBody Map<String, Object> body) {
        String bizNo = "SQP" + System.currentTimeMillis();
        String sql = "INSERT INTO sq_inspection (biz_no, building_id, room_no, score, hygiene_status, safety_status, inspector_name, remark, patrol_time, created_at, updated_at, is_deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)";
        LocalDateTime now = LocalDateTime.now();
        jdbcTemplate.update(sql, bizNo, body.get("building_id"), body.get("room_no"), body.get("score"), body.get("hygiene_status"), body.get("safety_status"), body.get("inspector_name"), body.get("remark"), body.get("patrol_time"), now, now);
        body.put("biz_no", bizNo);
        return R.ok(body);
    }

    @PostMapping("/buildings")
    public R<Map<String, Object>> createBuilding(@RequestBody Map<String, Object> body) {
        String sql = "INSERT INTO sq_building (code, name, total_floors, tutor_id, created_at, updated_at, is_deleted) VALUES (?, ?, ?, ?, ?, ?, 0)";
        LocalDateTime now = LocalDateTime.now();
        jdbcTemplate.update(sql, body.get("code"), body.get("name"), body.get("total_floors"), body.get("tutor_id"), now, now);
        return R.ok(body);
    }

    @PutMapping("/buildings/{id}")
    public R<Void> updateBuilding(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String sql = "UPDATE sq_building SET code = ?, name = ?, total_floors = ?, tutor_id = ?, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, body.get("code"), body.get("name"), body.get("total_floors"), body.get("tutor_id"), LocalDateTime.now(), id);
        return R.ok();
    }

    @DeleteMapping("/buildings/{id}")
    public R<Void> deleteBuilding(@PathVariable Long id) {
        String sql = "UPDATE sq_building SET is_deleted = 1, updated_at = ? WHERE id = ? AND is_deleted = 0";
        jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return R.ok();
    }
}
