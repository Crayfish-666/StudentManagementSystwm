package com.studenthub.modules.ty.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/ty")
public class TyModuleController {

    @GetMapping("/applications")
    public R<Map<String, Object>> getApplications() {
        List<Map<String, Object>> items = new ArrayList<>();
        items.add(createApp(1L, "TY-2026-0001", "张三", "2023010101", "计算机2301团支部", "计算机学院", "2026-03-01", "S3"));
        items.add(createApp(2L, "TY-2026-0002", "李四", "2023010102", "经管2302团支部", "经济管理学院", "2026-03-05", "S2"));
        items.add(createApp(3L, "TY-2026-0003", "王五", "2023010103", "艺术2301团支部", "艺术设计学院", "2026-03-10", "S1"));
        items.add(createApp(4L, "TY-2026-0004", "赵六", "2023010104", "软件2303团支部", "软件工程学院", "2026-03-12", "S0"));
        return R.ok(wrapPage(items));
    }

    @GetMapping("/approvals/pending")
    public R<Map<String, Object>> getPendingApprovals() {
        List<Map<String, Object>> items = new ArrayList<>();
        items.add(createApp(3L, "TY-2026-0003", "王五", "2023010103", "艺术2301团支部", "艺术设计学院", "2026-03-10", "S1"));
        return R.ok(wrapPage(items));
    }

    @GetMapping("/recommendation-meetings")
    public R<Map<String, Object>> getRecommendationMeetings() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> m = new HashMap<>();
        m.put("id", 101L);
        m.put("title", "2026年春季计算机学院推优大会");
        m.put("branch_name", "计算机2301团支部");
        m.put("meeting_date", "2026-03-15");
        m.put("attendee_rate", "95.2%");
        m.put("status", "completed");
        items.add(m);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/cultivation-links")
    public R<Map<String, Object>> getCultivationLinks() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> c = new HashMap<>();
        c.put("id", 201L);
        c.put("student_name", "张三");
        c.put("student_no", "2023010101");
        c.put("cultivator_name", "陈辅导员");
        c.put("stage", "培养联系期");
        c.put("progress", "80%");
        items.add(c);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/development-objects")
    public R<Map<String, Object>> getDevelopmentObjects() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> d = new HashMap<>();
        d.put("id", 301L);
        d.put("student_name", "张三");
        d.put("student_no", "2023010101");
        d.put("college_name", "计算机学院");
        d.put("approval_date", "2026-03-01");
        d.put("status", "active");
        items.add(d);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/political-reviews")
    public R<Map<String, Object>> getPoliticalReviews() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> p = new HashMap<>();
        p.put("id", 401L);
        p.put("student_name", "张三");
        p.put("student_no", "2023010101");
        p.put("result", "合格");
        p.put("reviewer", "张书记");
        p.put("review_date", "2026-03-18");
        items.add(p);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/development-meetings")
    public R<Map<String, Object>> getDevelopmentMeetings() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> m = new HashMap<>();
        m.put("id", 501L);
        m.put("meeting_title", "2026上半年团员接收发展大会");
        m.put("meeting_date", "2026-03-20");
        m.put("pass_count", 15);
        items.add(m);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/probationary-members")
    public R<Map<String, Object>> getProbationaryMembers() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> pr = new HashMap<>();
        pr.put("id", 601L);
        pr.put("student_name", "张三");
        pr.put("student_no", "2023010101");
        pr.put("probation_start", "2025-03-20");
        pr.put("probation_end", "2026-03-20");
        pr.put("status", "ready_for_regular");
        items.add(pr);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/branches")
    public R<Map<String, Object>> getBranches() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> b1 = new HashMap<>();
        b1.put("id", 1);
        b1.put("name", "计算机2301团支部");
        items.add(b1);
        Map<String, Object> b2 = new HashMap<>();
        b2.put("id", 2);
        b2.put("name", "经管2302团支部");
        items.add(b2);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/members")
    public R<Map<String, Object>> getMembers() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> m = new HashMap<>();
        m.put("id", 1L);
        m.put("student_name", "张三");
        m.put("student_no", "2023010101");
        m.put("league_no", "TY202600123");
        m.put("join_date", "2024-05-04");
        items.add(m);
        return R.ok(wrapPage(items));
    }

    private Map<String, Object> createApp(Long id, String bizNo, String sName, String sNo, String bName, String cName, String date, String status) {
        Map<String, Object> app = new HashMap<>();
        app.put("id", id);
        app.put("biz_no", bizNo);
        app.put("student_name", sName);
        app.put("student_no", sNo);
        app.put("branch_name", bName);
        app.put("college_name", cName);
        app.put("apply_date", date);
        app.put("status", status);
        return app;
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
