package com.studenthub.modules.sys.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;
import java.time.LocalDateTime;

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

    // --- 用户管理 (User Management) ---

    @PostMapping("/users")
    public R<Map<String, Object>> createUser(@RequestBody Map<String, Object> body) {
        String sql = "INSERT INTO sys_user (username, password_hash, display_name, user_type, status, student_id, created_at, updated_at, is_deleted) " +
                "VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)";
        String username = (String) body.get("username");
        String passwordHash = (String) body.get("password_hash");
        String displayName = (String) body.get("display_name");
        String userType = (String) body.getOrDefault("user_type", "student");
        String status = (String) body.getOrDefault("status", "active");
        Integer studentId = body.get("student_id") != null ? Integer.valueOf(body.get("student_id").toString()) : null;
        LocalDateTime now = LocalDateTime.now();

        jdbcTemplate.update(sql, username, passwordHash, displayName, userType, status, studentId, now, now);

        String fetchSql = "SELECT * FROM sys_user WHERE username = ? AND is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(fetchSql, username);
        if (items.isEmpty()) return R.fail(500, "创建用户失败");
        return R.ok(items.get(0));
    }

    @PutMapping("/users/{id}")
    public R<Void> updateUser(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String sql = "UPDATE sys_user SET display_name = ?, user_type = ?, status = ?, updated_at = ? WHERE id = ? AND is_deleted = 0";
        String displayName = (String) body.get("display_name");
        String userType = (String) body.get("user_type");
        String status = (String) body.get("status");
        
        int rows = jdbcTemplate.update(sql, displayName, userType, status, LocalDateTime.now(), id);
        return rows > 0 ? R.ok() : R.fail(404, "用户不存在");
    }

    @DeleteMapping("/users/{id}")
    public R<Void> deleteUser(@PathVariable Long id) {
        String sql = "UPDATE sys_user SET is_deleted = 1, updated_at = ? WHERE id = ? AND is_deleted = 0";
        int rows = jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return rows > 0 ? R.ok() : R.fail(404, "用户不存在");
    }

    @PostMapping("/users/{id}/reset-password")
    public R<Void> resetPassword(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String newPassword = (String) body.get("new_password");
        String sql = "UPDATE sys_user SET password_hash = ?, updated_at = ? WHERE id = ? AND is_deleted = 0";
        int rows = jdbcTemplate.update(sql, newPassword, LocalDateTime.now(), id);
        return rows > 0 ? R.ok() : R.fail(404, "用户不存在");
    }

    @PostMapping("/users/{id}/lock")
    public R<Void> lockUser(@PathVariable Long id) {
        String sql = "UPDATE sys_user SET status = 'locked', updated_at = ? WHERE id = ? AND is_deleted = 0";
        int rows = jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return rows > 0 ? R.ok() : R.fail(404, "用户不存在");
    }

    @PostMapping("/users/{id}/unlock")
    public R<Void> unlockUser(@PathVariable Long id) {
        String sql = "UPDATE sys_user SET status = 'active', updated_at = ? WHERE id = ? AND is_deleted = 0";
        int rows = jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return rows > 0 ? R.ok() : R.fail(404, "用户不存在");
    }

    @PostMapping("/users/{id}/enable")
    public R<Void> enableUser(@PathVariable Long id) {
        return unlockUser(id);
    }

    @PostMapping("/users/{id}/disable")
    public R<Void> disableUser(@PathVariable Long id) {
        String sql = "UPDATE sys_user SET status = 'disabled', updated_at = ? WHERE id = ? AND is_deleted = 0";
        int rows = jdbcTemplate.update(sql, LocalDateTime.now(), id);
        return rows > 0 ? R.ok() : R.fail(404, "用户不存在");
    }

    @PostMapping("/users/{id}/roles")
    public R<Void> assignRoles(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        List<?> roleIds = (List<?>) body.get("role_ids");
        if (roleIds == null) return R.fail(400, "需要 role_ids 参数");
        
        try {
            jdbcTemplate.update("DELETE FROM sys_user_role WHERE user_id = ?", id);
            for (Object roleId : roleIds) {
                jdbcTemplate.update("INSERT INTO sys_user_role (user_id, role_id, created_at) VALUES (?, ?, ?)", 
                        id, roleId, LocalDateTime.now());
            }
        } catch (Exception e) {
            return R.fail(500, "分配角色失败");
        }
        return R.ok();
    }

    @DeleteMapping("/users/{id}/roles/{roleId}")
    public R<Void> deleteUserRole(@PathVariable Long id, @PathVariable Long roleId) {
        try {
            jdbcTemplate.update("DELETE FROM sys_user_role WHERE user_id = ? AND role_id = ?", id, roleId);
        } catch (Exception e) {
            return R.fail(500, "移除角色失败");
        }
        return R.ok();
    }

    // --- 菜单管理 (Menu Management) ---

    @GetMapping("/menus/mine")
    public R<List<Map<String, Object>>> getMyMenus() {
        return getMenus();
    }

    // --- 字典管理 (Dictionary Management) ---

    @GetMapping("/dicts")
    public R<Map<String, Object>> getDicts(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size) {
        return getDictTypes(page, page_size);
    }

    @GetMapping("/dicts/{category}/items")
    public R<List<Map<String, Object>>> getDictItemsByCategory(@PathVariable String category) {
        return getDictByType(category);
    }

    @PostMapping("/dicts/items")
    public R<Map<String, Object>> createDictItem(@RequestBody Map<String, Object> body) {
        String sql = "INSERT INTO sys_dict (category, code, name_zh, sort, extra_json, is_active, is_deleted) " +
                "VALUES (?, ?, ?, ?, ?, 1, 0)";
        String category = (String) body.get("category");
        String code = (String) body.get("code");
        String nameZh = (String) body.get("name_zh");
        Integer sort = body.get("sort") != null ? Integer.valueOf(body.get("sort").toString()) : 0;
        String extraJson = (String) body.get("extra_json");
        
        jdbcTemplate.update(sql, category, code, nameZh, sort, extraJson);
        
        return R.ok(body);
    }

    @PutMapping("/dicts/items/{id}")
    public R<Void> updateDictItem(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        String sql = "UPDATE sys_dict SET category = ?, code = ?, name_zh = ?, sort = ?, extra_json = ? WHERE id = ? AND is_deleted = 0";
        String category = (String) body.get("category");
        String code = (String) body.get("code");
        String nameZh = (String) body.get("name_zh");
        Integer sort = body.get("sort") != null ? Integer.valueOf(body.get("sort").toString()) : 0;
        String extraJson = (String) body.get("extra_json");
        
        int rows = jdbcTemplate.update(sql, category, code, nameZh, sort, extraJson, id);
        return rows > 0 ? R.ok() : R.fail(404, "字典项不存在");
    }

    @DeleteMapping("/dicts/items/{id}")
    public R<Void> deleteDictItem(@PathVariable Long id) {
        String sql = "UPDATE sys_dict SET is_deleted = 1 WHERE id = ? AND is_deleted = 0";
        int rows = jdbcTemplate.update(sql, id);
        return rows > 0 ? R.ok() : R.fail(404, "字典项不存在");
    }
}
