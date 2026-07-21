package com.studenthub.modules.cmp.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/cmp")
public class CmpModuleController {

    @GetMapping("/rankings")
    public R<Map<String, Object>> getRankings() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> r1 = new HashMap<>();
        r1.put("rank", 1);
        r1.put("student_name", "张三");
        r1.put("student_no", "2023010101");
        r1.put("college_name", "计算机学院");
        r1.put("total_score", 94.8);
        items.add(r1);

        Map<String, Object> r2 = new HashMap<>();
        r2.put("rank", 2);
        r2.put("student_name", "李四");
        r2.put("student_no", "2023010102");
        r2.put("college_name", "经济管理学院");
        r2.put("total_score", 92.5);
        items.add(r2);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/scores")
    public R<Map<String, Object>> getScores() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> s = new HashMap<>();
        s.put("id", 1L);
        s.put("student_name", "张三");
        s.put("academic_score", 90.0);
        s.put("moral_score", 95.0);
        s.put("physical_score", 92.0);
        s.put("art_score", 88.0);
        s.put("labor_score", 96.0);
        s.put("total_score", 92.2);
        items.add(s);
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
