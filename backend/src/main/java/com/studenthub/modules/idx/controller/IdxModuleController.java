package com.studenthub.modules.idx.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;

/**
 * 学生身份库模块 Controller
 */
@RestController
@RequestMapping("/idx")
public class IdxModuleController {

    private final JdbcTemplate jdbcTemplate;

    public IdxModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/students")
    public R<Map<String, Object>> getStudents(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int page_size,
            @RequestParam(required = false) Integer college_id,
            @RequestParam(required = false) Integer class_id,
            @RequestParam(required = false) String status) {

        StringBuilder where = new StringBuilder("WHERE s.is_deleted = 0 ");
        List<Object> params = new ArrayList<>();

        if (college_id != null) {
            where.append("AND s.college_id = ? ");
            params.add(college_id);
        }
        if (class_id != null) {
            where.append("AND s.class_id = ? ");
            params.add(class_id);
        }
        if (status != null && !status.trim().isEmpty()) {
            where.append("AND s.status = ? ");
            params.add(status);
        }

        String countSql = "SELECT COUNT(*) FROM idx_student s " + where;
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class, params.toArray());
        if (total == null) total = 0;

        String sql = "SELECT s.id, s.student_no, s.name, s.name as student_name, s.gender, " +
                "s.political_status, s.birth_date, s.status, " +
                "col.id as college_id, col.name as college_name, " +
                "m.id as major_id, m.name as major_name, " +
                "c.id as class_id, c.name as class_name, c.grade, " +
                "s.created_at, s.updated_at " +
                "FROM idx_student s " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "LEFT JOIN sys_major m ON s.major_id = m.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                where +
                "ORDER BY s.student_no ASC " +
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

    @GetMapping("/students/{id}")
    public R<Map<String, Object>> getStudent(@PathVariable Long id) {
        String sql = "SELECT s.id, s.student_no, s.name, s.gender, s.political_status, " +
                "s.birth_date, s.status, " +
                "col.id as college_id, col.name as college_name, " +
                "m.id as major_id, m.name as major_name, " +
                "c.id as class_id, c.name as class_name, c.grade, " +
                "s.created_at, s.updated_at " +
                "FROM idx_student s " +
                "LEFT JOIN sys_college col ON s.college_id = col.id " +
                "LEFT JOIN sys_major m ON s.major_id = m.id " +
                "LEFT JOIN idx_class c ON s.class_id = c.id " +
                "WHERE s.id = ? AND s.is_deleted = 0";
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql, id);
        if (rows.isEmpty()) {
            return R.fail(1404, "学生不存在");
        }
        return R.ok(rows.get(0));
    }

    @GetMapping("/org-tree")
    public R<List<Map<String, Object>>> getOrgTree() {
        // 构建 学院 -> 专业 -> 班级 三级树
        String collegeSql = "SELECT id, code, name FROM sys_college WHERE is_deleted = 0 ORDER BY id";
        List<Map<String, Object>> colleges = jdbcTemplate.queryForList(collegeSql);

        String majorSql = "SELECT id, college_id, code, name FROM sys_major WHERE is_deleted = 0 ORDER BY id";
        List<Map<String, Object>> majors = jdbcTemplate.queryForList(majorSql);

        String classSql = "SELECT id, major_id, grade, code, name FROM idx_class WHERE is_deleted = 0 ORDER BY grade, id";
        List<Map<String, Object>> classes = jdbcTemplate.queryForList(classSql);

        List<Map<String, Object>> tree = new ArrayList<>();
        for (Map<String, Object> college : colleges) {
            Long colId = ((Number) college.get("id")).longValue();
            Map<String, Object> node = new HashMap<>();
            node.put("id", colId);
            node.put("code", college.get("code"));
            node.put("name", college.get("name"));
            node.put("type", "college");

            List<Map<String, Object>> majorChildren = new ArrayList<>();
            for (Map<String, Object> major : majors) {
                Long majColId = major.get("college_id") != null ?
                        ((Number) major.get("college_id")).longValue() : null;
                if (majColId != null && majColId.equals(colId)) {
                    Long majId = ((Number) major.get("id")).longValue();
                    Map<String, Object> majorNode = new HashMap<>();
                    majorNode.put("id", majId);
                    majorNode.put("code", major.get("code"));
                    majorNode.put("name", major.get("name"));
                    majorNode.put("type", "major");

                    List<Map<String, Object>> classChildren = new ArrayList<>();
                    for (Map<String, Object> cls : classes) {
                        Long clsMajId = cls.get("major_id") != null ?
                                ((Number) cls.get("major_id")).longValue() : null;
                        if (clsMajId != null && clsMajId.equals(majId)) {
                            Map<String, Object> classNode = new HashMap<>();
                            classNode.put("id", cls.get("id"));
                            classNode.put("code", cls.get("code"));
                            classNode.put("name", cls.get("name"));
                            classNode.put("grade", cls.get("grade"));
                            classNode.put("type", "class");
                            classChildren.add(classNode);
                        }
                    }
                    majorNode.put("children", classChildren);
                    majorChildren.add(majorNode);
                }
            }
            node.put("children", majorChildren);
            tree.add(node);
        }
        return R.ok(tree);
    }

    @GetMapping("/profile/me")
    public R<Map<String, Object>> getMyProfile() {
        try {
            Object loginId = cn.dev33.satoken.stp.StpUtil.getLoginId();
            if (loginId == null) return R.fail(10401, "未登录");
            String username = String.valueOf(loginId);

            // 先查 sys_user
            List<Map<String, Object>> userRows = jdbcTemplate.queryForList(
                    "SELECT id, username, display_name, user_type, status, student_id, last_login_at " +
                            "FROM sys_user WHERE username = ? AND is_deleted = 0", username);
            if (userRows.isEmpty()) {
                return R.fail(1404, "用户不存在");
            }
            Map<String, Object> user = userRows.get(0);

            // 如果关联了学生，查学生详情
            Object studentId = user.get("student_id");
            if (studentId != null) {
                String sql = "SELECT s.id, s.student_no, s.name, s.gender, s.political_status, " +
                        "s.birth_date, s.status, " +
                        "col.name as college_name, m.name as major_name, c.name as class_name, c.grade " +
                        "FROM idx_student s " +
                        "LEFT JOIN sys_college col ON s.college_id = col.id " +
                        "LEFT JOIN sys_major m ON s.major_id = m.id " +
                        "LEFT JOIN idx_class c ON s.class_id = c.id " +
                        "WHERE s.id = ? AND s.is_deleted = 0";
                List<Map<String, Object>> stuRows = jdbcTemplate.queryForList(sql, studentId);
                if (!stuRows.isEmpty()) {
                    user.put("student", stuRows.get(0));
                }
            }

            // 查角色
            List<Map<String, Object>> roles = jdbcTemplate.queryForList(
                    "SELECT r.id, r.code, r.name, r.scope FROM sys_role r " +
                            "INNER JOIN sys_user_role ur ON r.id = ur.role_id " +
                            "WHERE ur.user_id = ? AND r.is_deleted = 0",
                    ((Number) user.get("id")).longValue());
            user.put("roles", roles);

            return R.ok(user);
        } catch (Exception e) {
            return R.fail(10401, "未登录");
        }
    }
}
