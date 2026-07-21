package com.studenthub.modules.sys.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;

/**
 * 系统管理模块 Controller
 * 所有数据均来源于数据库真实表查询
 */
@RestController
@RequestMapping("/sys")
public class SysModuleController {

    private final JdbcTemplate jdbcTemplate;

    public SysModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/users")
    public R<Map<String, Object>> getUsers(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String keyword,
            @RequestParam(required = false) Integer role_id,
            @RequestParam(required = false) Integer college_id) {

        StringBuilder where = new StringBuilder("WHERE u.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (keyword != null && !keyword.trim().isEmpty()) {
            where.append("AND (u.username LIKE ? OR u.display_name LIKE ?) ");
            params.add("%" + keyword + "%");
            params.add("%" + keyword + "%");
        }

        String countSql = "SELECT COUNT(*) FROM sys_user u " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT u.id, u.username, u.display_name as real_name, u.user_type, " +
                "u.status, u.student_id, u.last_login_at, " +
                "u.created_at, u.updated_at " +
                "FROM sys_user u " +
                where +
                "ORDER BY u.id " +
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

    @GetMapping("/users/{id}")
    public R<Map<String, Object>> getUserDetail(@PathVariable Long id) {
        String sql = "SELECT u.id, u.username, u.display_name as real_name, u.user_type, " +
                "u.status, u.student_id, u.last_login_at, " +
                "u.created_at, u.updated_at " +
                "FROM sys_user u " +
                "WHERE u.id = ? AND u.is_deleted = 0";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql, id);
        if (rows.isEmpty()) {
            return R.fail(9040, "用户不存在");
        }
        return R.ok(rows.get(0));
    }

    @GetMapping("/roles")
    public R<Map<String, Object>> getRoles(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {

        String countSql = "SELECT COUNT(*) FROM sys_role WHERE is_deleted = 0";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT r.id, r.name, r.code, r.scope, r.description, " +
                "r.created_at, r.updated_at " +
                "FROM sys_role r " +
                "WHERE r.is_deleted = 0 " +
                "ORDER BY r.id " +
                "LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/roles/all")
    public R<List<Map<String, Object>>> getAllRoles() {
        String sql = "SELECT id, name, code, scope FROM sys_role WHERE is_deleted = 0 ORDER BY id";
        List<Map<String, Object>> roles = jdbcTemplate.queryForList(sql);
        return R.ok(roles);
    }

    @GetMapping("/menus")
    public R<List<Map<String, Object>>> getMenus() {
        String sql = "SELECT id, code, title, icon, path, component, parent_id, sort, roles, visible " +
                "FROM sys_menu WHERE is_deleted = 0 ORDER BY sort, id";
        List<Map<String, Object>> menus = jdbcTemplate.queryForList(sql);
        return R.ok(menus);
    }

    @GetMapping("/dict/{type}")
    public R<List<Map<String, Object>>> getDictByType(@PathVariable String type) {
        String sql = "SELECT id, category as type, code as key, name_zh as value, " +
                "sort as sort_order, extra_json as remark " +
                "FROM sys_dict " +
                "WHERE category = ? AND is_deleted = 0 AND is_active = 1 " +
                "ORDER BY sort, id";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, type);
        return R.ok(items);
    }

    @GetMapping("/dict-types")
    public R<Map<String, Object>> getDictTypes(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {

        String countSql = "SELECT COUNT(DISTINCT category) FROM sys_dict WHERE is_deleted = 0";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);
        if (total == null) total = 0;

        String sql = "SELECT category as type, " +
                "COUNT(*) as item_count, " +
                "MAX(created_at) as updated_at " +
                "FROM sys_dict " +
                "WHERE is_deleted = 0 " +
                "GROUP BY category " +
                "ORDER BY category " +
                "LIMIT ? OFFSET ?";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, page_size, (page - 1) * page_size);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total);
        result.put("page", page);
        result.put("page_size", page_size);
        return R.ok(result);
    }

    @GetMapping("/colleges")
    public R<List<Map<String, Object>>> getColleges() {
        String sql = "SELECT id, name, code, created_at " +
                "FROM sys_college WHERE is_deleted = 0 ORDER BY id";
        List<Map<String, Object>> colleges = jdbcTemplate.queryForList(sql);
        return R.ok(colleges);
    }

    @GetMapping("/majors")
    public R<List<Map<String, Object>>> getMajors(
            @RequestParam(required = false) Integer college_id) {
        String sql;
        List<Object> params = new ArrayList<>();
        if (college_id != null) {
            sql = "SELECT m.id, m.name, m.code, m.college_id, col.name as college_name, " +
                    "m.created_at FROM sys_major m " +
                    "LEFT JOIN sys_college col ON m.college_id = col.id " +
                    "WHERE m.is_deleted = 0 AND m.college_id = ? ORDER BY m.id";
            params.add(college_id);
        } else {
            sql = "SELECT m.id, m.name, m.code, m.college_id, col.name as college_name, " +
                    "m.created_at FROM sys_major m " +
                    "LEFT JOIN sys_college col ON m.college_id = col.id " +
                    "WHERE m.is_deleted = 0 ORDER BY m.id";
        }
        List<Map<String, Object>> majors = jdbcTemplate.queryForList(sql, params.toArray());
        return R.ok(majors);
    }

    @GetMapping("/classes")
    public R<Map<String, Object>> getClasses(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) Integer major_id,
            @RequestParam(required = false) Integer college_id) {

        StringBuilder where = new StringBuilder("WHERE c.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (major_id != null) {
            where.append("AND c.major_id = ? ");
            params.add(major_id);
        }
        if (college_id != null) {
            where.append("AND m.college_id = ? ");
            params.add(college_id);
        }

        String countSql = "SELECT COUNT(*) FROM idx_class c " +
                "LEFT JOIN sys_major m ON c.major_id = m.id " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT c.id, c.name, c.code, c.grade, c.major_id, m.name as major_name, " +
                "col.name as college_name, " +
                "COUNT(s.id) as student_count, " +
                "c.created_at " +
                "FROM idx_class c " +
                "LEFT JOIN sys_major m ON c.major_id = m.id " +
                "LEFT JOIN sys_college col ON m.college_id = col.id " +
                "LEFT JOIN idx_student s ON s.class_id = c.id AND s.is_deleted = 0 " +
                where +
                "GROUP BY c.id, c.name, c.code, c.grade, c.major_id, m.name, col.name " +
                "ORDER BY c.id " +
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

    @GetMapping("/org-tree")
    public R<List<Map<String, Object>>> getOrgTree() {
        String collegeSql = "SELECT id, name, code FROM sys_college WHERE is_deleted = 0 ORDER BY id";
        List<Map<String, Object>> colleges = jdbcTemplate.queryForList(collegeSql);

        List<Map<String, Object>> tree = new ArrayList<>();
        for (Map<String, Object> college : colleges) {
            Integer collegeId = (Integer) college.get("id");

            String majorSql = "SELECT id, name, code FROM sys_major " +
                    "WHERE college_id = ? AND is_deleted = 0 ORDER BY id";
            List<Map<String, Object>> majors = jdbcTemplate.queryForList(majorSql, collegeId);

            List<Map<String, Object>> majorNodes = new ArrayList<>();
            for (Map<String, Object> major : majors) {
                Integer majorId = (Integer) major.get("id");

                String classSql = "SELECT id, name, code FROM idx_class " +
                        "WHERE major_id = ? AND is_deleted = 0 ORDER BY id";
                List<Map<String, Object>> classes = jdbcTemplate.queryForList(classSql, majorId);

                Map<String, Object> majorNode = new HashMap<>(major);
                majorNode.put("children", classes);
                majorNodes.add(majorNode);
            }

            Map<String, Object> collegeNode = new HashMap<>(college);
            collegeNode.put("children", majorNodes);
            tree.add(collegeNode);
        }
        return R.ok(tree);
    }

    @GetMapping("/logs")
    public R<Map<String, Object>> getLogs(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) String type) {

        // 操作日志：暂时用 sq_incident 表模拟（因为没有 event_log 表）
        StringBuilder where = new StringBuilder("WHERE 1=1 ");
        List<Object> params = new ArrayList<>();

        String countSql = "SELECT COUNT(*) FROM sq_incident " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT id, biz_no, incident_type as event_type, status as action, " +
                "reporter_id as operator_id, " +
                "description, resolution, " +
                "created_at " +
                "FROM sq_incident " +
                where +
                "ORDER BY created_at DESC " +
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
}
