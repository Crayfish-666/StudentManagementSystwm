package com.studenthub.modules.st.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/st")
public class StModuleController {

    @GetMapping("/associations")
    public R<Map<String, Object>> getAssociations() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> a = new HashMap<>();
        a.put("id", 1L);
        a.put("name", "计算机算法与编程社团");
        a.put("category", "学术科技类");
        a.put("president_name", "张三");
        a.put("member_count", 128);
        a.put("star_level", 5);
        items.add(a);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/recruit-plans")
    public R<Map<String, Object>> getRecruitPlans() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> r = new HashMap<>();
        r.put("id", 1L);
        r.put("title", "2026春季招新计划");
        r.put("association_name", "计算机算法与编程社团");
        r.put("target_count", 50);
        r.put("applied_count", 32);
        r.put("status", "recruiting");
        items.add(r);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/recruit-applies")
    public R<Map<String, Object>> getRecruitApplies() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> app = new HashMap<>();
        app.put("id", 1L);
        app.put("student_name", "李四");
        app.put("student_no", "2023010102");
        app.put("association_name", "计算机算法与编程社团");
        app.put("status", "pending");
        items.add(app);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/activities")
    public R<Map<String, Object>> getActivities() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> act = new HashMap<>();
        act.put("id", 1L);
        act.put("title", "第十二届全校程序设计大赛");
        act.put("association_name", "计算机算法与编程社团");
        act.put("activity_date", "2026-04-10");
        act.put("location", "实验楼401机房");
        act.put("status", "approved");
        items.add(act);
        return R.ok(wrapPage(items));
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
