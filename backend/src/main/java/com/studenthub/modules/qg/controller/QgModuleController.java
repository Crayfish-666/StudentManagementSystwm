package com.studenthub.modules.qg.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/qg")
public class QgModuleController {

    @GetMapping("/difficulties")
    public R<Map<String, Object>> getDifficulties() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> d = new HashMap<>();
        d.put("id", 1L);
        d.put("student_name", "李四");
        d.put("student_no", "2023010102");
        d.put("level", "特殊困难");
        d.put("status", "approved");
        items.add(d);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/positions")
    public R<Map<String, Object>> getPositions() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> p = new HashMap<>();
        p.put("id", 1L);
        p.put("title", "图书馆图书整理助理");
        p.put("department", "图书馆");
        p.put("hourly_rate", 22.5);
        p.put("quota", 10);
        items.add(p);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/attendances")
    public R<Map<String, Object>> getAttendances() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> a = new HashMap<>();
        a.put("id", 1L);
        a.put("student_name", "李四");
        a.put("position_title", "图书馆图书整理助理");
        a.put("work_date", "2026-03-18");
        a.put("hours", 4.0);
        a.put("status", "approved");
        items.add(a);
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
