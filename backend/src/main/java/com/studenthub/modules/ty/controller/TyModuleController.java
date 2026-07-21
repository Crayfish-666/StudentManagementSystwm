package com.studenthub.modules.ty.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/ty")
public class TyModuleController {

    private final JdbcTemplate jdbcTemplate;

    public TyModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    private static final String[] NAMES = {
            "张伟", "王芳", "李娜", "刘洋", "陈杰",
            "杨光", "黄磊", "周敏", "吴强", "徐霞",
            "孙浩", "胡婷", "朱勇", "高丽", "林涛",
            "何静", "郭平", "马明", "罗军", "梁晨"
    };

    private static final String[] COLLEGES = {
            "计算机学院", "经济管理学院", "艺术设计学院", "软件工程学院", "电子信息工程学院"
    };

    private static final String[] BRANCHES = {
            "计算机2301团支部", "经管2302团支部", "艺术2301团支部", "软件2303团支部", "电信2302团支部"
    };

    @GetMapping("/applications")
    public R<Map<String, Object>> getApplications(@RequestParam(defaultValue = "1") int page,
                                                 @RequestParam(defaultValue = "20") int page_size) {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> app = new HashMap<>();
            app.put("id", (long) i);
            app.put("biz_no", String.format("TY-2026-%04d", i));
            app.put("student_name", NAMES[(i - 1) % NAMES.length]);
            app.put("student_no", String.format("20230101%02d", i));
            app.put("branch_name", BRANCHES[(i - 1) % BRANCHES.length]);
            app.put("college_name", COLLEGES[(i - 1) % COLLEGES.length]);
            app.put("apply_date", String.format("2026-03-%02d", (i % 28) + 1));
            app.put("status", i % 4 == 0 ? "S3" : (i % 3 == 0 ? "S2" : (i % 2 == 0 ? "S1" : "S0")));
            app.put("created_at", String.format("2026-03-%02d 10:00:00", (i % 28) + 1));
            items.add(app);
        }
        return R.ok(wrapPage(items, page, page_size));
    }

    @GetMapping("/approvals/pending")
    public R<Map<String, Object>> getPendingApprovals() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> app = new HashMap<>();
            app.put("id", (long) i);
            app.put("biz_no", String.format("TY-2026-%04d", i));
            app.put("student_name", NAMES[(i - 1) % NAMES.length]);
            app.put("student_no", String.format("20230101%02d", i));
            app.put("branch_name", BRANCHES[(i - 1) % BRANCHES.length]);
            app.put("college_name", COLLEGES[(i - 1) % COLLEGES.length]);
            app.put("apply_date", String.format("2026-03-%02d", (i % 28) + 1));
            app.put("status", "S1");
            app.put("created_at", String.format("2026-03-%02d 09:30:00", (i % 28) + 1));
            items.add(app);
        }
        return R.ok(wrapPage(items, 1, 20));
    }

    @GetMapping("/recommendation-meetings")
    public R<Map<String, Object>> getRecommendationMeetings() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> m = new HashMap<>();
            m.put("id", (long) i);
            m.put("title", String.format("2026春季%s推优大会第%d期", BRANCHES[(i - 1) % BRANCHES.length], i));
            m.put("branch_name", BRANCHES[(i - 1) % BRANCHES.length]);
            m.put("meeting_date", String.format("2026-03-%02d", (i % 28) + 1));
            m.put("attendee_rate", String.format("%.1f%%", 90.0 + (i % 10)));
            m.put("status", i % 2 == 0 ? "completed" : "scheduled");
            m.put("created_at", String.format("2026-03-%02d 14:00:00", (i % 28) + 1));
            items.add(m);
        }
        return R.ok(wrapPage(items, 1, 20));
    }

    @GetMapping("/cultivation-links")
    public R<Map<String, Object>> getCultivationLinks() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> c = new HashMap<>();
            c.put("id", (long) i);
            c.put("student_name", NAMES[(i - 1) % NAMES.length]);
            c.put("student_no", String.format("20230101%02d", i));
            c.put("cultivator_name", "张辅导员");
            c.put("stage", i % 2 == 0 ? "考察期" : "培养联系期");
            c.put("progress", String.format("%d%%", 50 + i * 2));
            c.put("created_at", String.format("2026-03-%02d 11:00:00", (i % 28) + 1));
            items.add(c);
        }
        return R.ok(wrapPage(items, 1, 20));
    }

    @GetMapping("/development-objects")
    public R<Map<String, Object>> getDevelopmentObjects() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> d = new HashMap<>();
            d.put("id", (long) i);
            d.put("biz_no", String.format("TY-DEV-2026-%04d", i));
            d.put("student_name", NAMES[(i - 1) % NAMES.length]);
            d.put("student_no", String.format("20230101%02d", i));
            d.put("branch_name", BRANCHES[(i - 1) % BRANCHES.length]);
            d.put("college_name", COLLEGES[(i - 1) % COLLEGES.length]);
            d.put("approval_date", String.format("2026-02-%02d", (i % 20) + 1));
            d.put("status", i % 3 == 0 ? "S4" : (i % 2 == 0 ? "S2" : "S1"));
            d.put("status_text", i % 3 == 0 ? "已通过" : (i % 2 == 0 ? "待审批" : "公示中"));
            d.put("created_at", String.format("2026-03-%02d 15:30:00", (i % 28) + 1));
            items.add(d);
        }
        return R.ok(wrapPage(items, 1, 20));
    }

    @GetMapping("/political-reviews")
    public R<Map<String, Object>> getPoliticalReviews() {
        List<Map<String, Object>> items = new ArrayList<>();
        String[] relations = { "father", "mother", "brother", "sister" };
        String[] conclusions = { "pass", "basic_pass", "fail" };
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> p = new HashMap<>();
            p.put("id", (long) i);
            p.put("target_name", NAMES[(i - 1) % NAMES.length] + " (家属)");
            p.put("target_relation", relations[(i - 1) % relations.length]);
            p.put("method", i % 2 == 0 ? "letter" : "interview");
            p.put("conclusion", conclusions[(i - 1) % conclusions.length]);
            p.put("document_path", String.format("/storage/political_review_%03d.pdf", i));
            p.put("is_extend_3m", i % 5 == 0 ? 1 : 0);
            p.put("created_at", String.format("2026-03-%02d 16:45:00", (i % 25) + 1));
            items.add(p);
        }
        return R.ok(wrapPage(items, 1, 20));
    }

    @GetMapping("/development-meetings")
    public R<Map<String, Object>> getDevelopmentMeetings() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> m = new HashMap<>();
            m.put("id", (long) i);
            m.put("meeting_title", String.format("2026上半年%s预备团员接收大会", COLLEGES[(i - 1) % COLLEGES.length]));
            m.put("meeting_date", String.format("2026-03-%02d", (i % 25) + 1));
            m.put("pass_count", 10 + i);
            m.put("created_at", String.format("2026-03-%02d 09:00:00", (i % 25) + 1));
            items.add(m);
        }
        return R.ok(wrapPage(items, 1, 20));
    }

    @GetMapping("/probationary-members")
    public R<Map<String, Object>> getProbationaryMembers() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> pr = new HashMap<>();
            pr.put("id", (long) i);
            pr.put("student_name", NAMES[(i - 1) % NAMES.length]);
            pr.put("student_no", String.format("20230101%02d", i));
            pr.put("probation_start", "2025-03-20");
            pr.put("probation_end", "2026-03-20");
            pr.put("status", i % 3 == 0 ? "ready_for_regular" : "in_probation");
            pr.put("created_at", String.format("2026-03-%02d 10:20:00", (i % 25) + 1));
            items.add(pr);
        }
        return R.ok(wrapPage(items, 1, 20));
    }

    @GetMapping("/branches")
    public R<Map<String, Object>> getBranches() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> b = new HashMap<>();
            b.put("id", (long) i);
            b.put("name", String.format("2023级%s团支部第%d分部", COLLEGES[(i - 1) % COLLEGES.length], i));
            items.add(b);
        }
        return R.ok(wrapPage(items, 1, 20));
    }

    @GetMapping("/members")
    public R<Map<String, Object>> getMembers() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> m = new HashMap<>();
            m.put("id", (long) i);
            m.put("biz_no", String.format("TY2026%05d", i * 11));
            m.put("student_no", String.format("20230101%02d", i));
            m.put("student_name", NAMES[(i - 1) % NAMES.length]);
            m.put("branch_name", BRANCHES[(i - 1) % BRANCHES.length]);
            m.put("join_at", String.format("2024-05-%02d", (i % 28) + 1));
            m.put("become_probationary_at", String.format("2025-05-%02d", (i % 28) + 1));
            m.put("formal_join_at", String.format("2026-05-%02d", (i % 28) + 1));
            m.put("status", i % 4 == 0 ? "transferred" : "active");
            m.put("created_at", String.format("2026-03-%02d 14:15:00", (i % 28) + 1));
            items.add(m);
        }
        return R.ok(wrapPage(items, 1, 20));
    }

    private Map<String, Object> wrapPage(List<Map<String, Object>> items, int page, int pageSize) {
        Map<String, Object> res = new HashMap<>();
        res.put("items", items);
        res.put("total", items.size());
        res.put("page", page);
        res.put("page_size", pageSize);
        return res;
    }
}
