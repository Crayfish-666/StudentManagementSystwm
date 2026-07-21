package com.studenthub.modules.sys.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;

@RestController
@RequestMapping("/sys")
public class SysModuleController {

    private final JdbcTemplate jdbcTemplate;

    public SysModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @GetMapping("/users")
    public R<Map<String, Object>> getUsers() {
        String countSql = "SELECT COUNT(*) FROM sys_user WHERE is_deleted = 0";
        Integer total = jdbcTemplate.queryForObject(countSql, Integer.class);

        String sql = "SELECT id, username, display_name, user_type, status, created_at " +
                     "FROM sys_user WHERE is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql);

        Map<String, Object> result = new HashMap<>();
        result.put("items", items);
        result.put("total", total != null ? total : items.size());
        result.put("page", 1);
        result.put("page_size", 20);
        return R.ok(result);
    }

    @GetMapping("/roles")
    public R<List<Map<String, Object>>> getRoles() {
        String sql = "SELECT id, code, name, scope FROM sys_role WHERE is_deleted = 0";
        List<Map<String, Object>> roles = jdbcTemplate.queryForList(sql);
        return R.ok(roles);
    }

    @GetMapping("/dicts")
    public R<List<Map<String, Object>>> getDictCategories() {
        String sql = "SELECT id, category, code, name_zh, sort, is_active FROM sys_dict WHERE is_deleted = 0";
        List<Map<String, Object>> dicts = jdbcTemplate.queryForList(sql);
        return R.ok(dicts);
    }

    @GetMapping("/dicts/{category}/items")
    public R<List<Map<String, Object>>> getDictItems(@PathVariable String category) {
        String sql = "SELECT id, category, code, name_zh, sort, is_active FROM sys_dict WHERE category = ? AND is_deleted = 0";
        List<Map<String, Object>> items = jdbcTemplate.queryForList(sql, category);
        return R.ok(items);
    }

    @GetMapping("/orgs")
    public R<List<Map<String, Object>>> getOrgs() {
        List<Map<String, Object>> orgs = new ArrayList<>();
        Map<String, Object> root = new HashMap<>();
        root.put("id", 1L);
        root.put("code", "ORG-01");
        root.put("name", "StudentHub Campus Administration");
        root.put("parent_id", null);
        orgs.add(root);
        return R.ok(orgs);
    }

    @GetMapping("/jobs")
    public R<List<Map<String, Object>>> getJobs() {
        List<Map<String, Object>> jobs = new ArrayList<>();
        Map<String, Object> job = new HashMap<>();
        job.put("id", 1L);
        job.put("job_name", "DataSyncJob");
        job.put("cron_expression", "0 0 2 * * ?");
        job.put("status", "NORMAL");
        job.put("last_run_at", "2026-07-22 02:00:00");
        jobs.add(job);
        return R.ok(jobs);
    }
}
