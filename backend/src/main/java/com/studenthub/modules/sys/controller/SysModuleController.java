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

    private static final String[] NAMES = {
            "张伟", "王芳", "李娜", "刘洋", "陈杰",
            "杨光", "黄磊", "周敏", "吴强", "徐霞",
            "孙浩", "胡婷", "朱勇", "高丽", "林涛",
            "何静", "郭平", "马明", "罗军", "梁晨"
    };

    @GetMapping("/users")
    public R<Map<String, Object>> getUsers() {
        List<Map<String, Object>> items = new ArrayList<>();

        Map<String, Object> admin = new HashMap<>();
        admin.put("id", 1L);
        admin.put("username", "admin");
        admin.put("display_name", "超级管理员");
        admin.put("user_type", "school");
        admin.put("status", "active");
        admin.put("created_at", "2026-03-01 08:00:00");
        items.add(admin);

        Map<String, Object> coun = new HashMap<>();
        coun.put("id", 2L);
        coun.put("username", "counselor");
        coun.put("display_name", "张辅导员");
        coun.put("user_type", "college");
        coun.put("status", "active");
        coun.put("created_at", "2026-03-01 08:30:00");
        items.add(coun);

        for (int i = 1; i <= 20; i++) {
            Map<String, Object> u = new HashMap<>();
            u.put("id", (long) (i + 2));
            u.put("username", String.format("20230101%02d", i));
            u.put("display_name", NAMES[i - 1]);
            u.put("user_type", "student");
            u.put("status", "active");
            u.put("created_at", String.format("2026-03-%02d 09:00:00", (i % 25) + 1));
            items.add(u);
        }
        return R.ok(wrapPage(items));
    }

    @GetMapping("/roles")
    public R<List<Map<String, Object>>> getRoles() {
        List<Map<String, Object>> roles = new ArrayList<>();
        roles.add(createRole(1L, "R-SY-ADMIN", "超级管理员", "school"));
        roles.add(createRole(2L, "R-SY-LEAGUE", "校团委管理员", "school"));
        roles.add(createRole(3L, "R-SY-AFFAIRS", "学工处管理员", "school"));
        roles.add(createRole(4L, "R-COL-LEAGUE", "院系团委", "college"));
        roles.add(createRole(5L, "R-COL-COUN", "院系辅导员", "college"));
        roles.add(createRole(6L, "R-STU-NORM", "普通学生", "student"));
        roles.add(createRole(7L, "R-STU-LEAGUE", "团支书", "student"));
        roles.add(createRole(8L, "R-STU-ASSOC", "社长/干事", "student"));
        roles.add(createRole(9L, "R-STU-COMMUNITY", "楼层长/舍长", "student"));
        return R.ok(roles);
    }

    @GetMapping("/dicts")
    public R<List<Map<String, Object>>> getDictCategories() {
        List<Map<String, Object>> dicts = new ArrayList<>();
        dicts.add(createDict(1L, "POLITICAL_STATUS", "PARTY_MEMBER", "中共党员", 1));
        dicts.add(createDict(2L, "POLITICAL_STATUS", "YOUTH_LEAGUE", "共青团员", 2));
        dicts.add(createDict(3L, "POLITICAL_STATUS", "MASSES", "群众", 3));
        dicts.add(createDict(4L, "TY_STATUS", "S0", "待提交", 1));
        dicts.add(createDict(5L, "TY_STATUS", "S1", "公示中", 2));
        dicts.add(createDict(6L, "TY_STATUS", "S2", "待审批", 3));
        dicts.add(createDict(7L, "TY_STATUS", "S3", "政审中", 4));
        dicts.add(createDict(8L, "TY_STATUS", "S4", "已通过", 5));
        return R.ok(dicts);
    }

    @GetMapping("/dicts/{category}/items")
    public R<List<Map<String, Object>>> getDictItems(@PathVariable String category) {
        List<Map<String, Object>> items = new ArrayList<>();
        if ("POLITICAL_STATUS".equalsIgnoreCase(category)) {
            items.add(createDict(1L, "POLITICAL_STATUS", "PARTY_MEMBER", "中共党员", 1));
            items.add(createDict(2L, "POLITICAL_STATUS", "YOUTH_LEAGUE", "共青团员", 2));
            items.add(createDict(3L, "POLITICAL_STATUS", "MASSES", "群众", 3));
        } else {
            items.add(createDict(4L, category, "S0", "待提交", 1));
            items.add(createDict(5L, category, "S1", "已提交", 2));
            items.add(createDict(6L, category, "S2", "审批中", 3));
            items.add(createDict(7L, category, "S3", "通过", 4));
        }
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

    private Map<String, Object> createRole(Long id, String code, String name, String scope) {
        Map<String, Object> r = new HashMap<>();
        r.put("id", id);
        r.put("code", code);
        r.put("name", name);
        r.put("scope", scope);
        return r;
    }

    private Map<String, Object> createDict(Long id, String category, String code, String nameZh, int sort) {
        Map<String, Object> d = new HashMap<>();
        d.put("id", id);
        d.put("category", category);
        d.put("code", code);
        d.put("name_zh", nameZh);
        d.put("sort", sort);
        d.put("is_active", 1);
        return d;
    }

    private Map<String, Object> wrapPage(List<Map<String, Object>> items) {
        Map<String, Object> res = new HashMap<>();
        res.put("items", items);
        res.put("total", items.size());
        res.put("page", 1);
        res.put("page_size", 20);
        return res;
    }
}
