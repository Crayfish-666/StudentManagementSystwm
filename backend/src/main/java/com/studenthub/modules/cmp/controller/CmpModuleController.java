package com.studenthub.modules.cmp.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/cmp")
public class CmpModuleController {

    private static final String[] NAMES = {
            "张伟", "王芳", "李娜", "刘洋", "陈杰",
            "杨光", "黄磊", "周敏", "吴强", "徐霞",
            "孙浩", "胡婷", "朱勇", "高丽", "林涛",
            "何静", "郭平", "马明", "罗军", "梁晨"
    };

    private static final String[] COLLEGES = {
            "计算机学院", "经济管理学院", "艺术设计学院", "软件工程学院", "电子信息工程学院"
    };

    @GetMapping("/rankings")
    public R<Map<String, Object>> getRankings() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> r = new HashMap<>();
            r.put("rank", i);
            r.put("student_name", NAMES[i - 1]);
            r.put("student_no", String.format("20230101%02d", i));
            r.put("college_name", COLLEGES[(i - 1) % COLLEGES.length]);
            r.put("total_score", String.format("%.2f", 98.0 - (i * 0.85)));
            items.add(r);
        }
        return R.ok(wrapPage(items));
    }

    @GetMapping("/scores")
    public R<Map<String, Object>> getScores() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> s = new HashMap<>();
            s.put("id", (long) i);
            s.put("student_name", NAMES[i - 1]);
            s.put("academic_score", 85.0 + (i % 10));
            s.put("moral_score", 90.0 + (i % 8));
            s.put("physical_score", 88.0 + (i % 9));
            s.put("art_score", 85.0 + (i % 12));
            s.put("labor_score", 92.0 + (i % 6));
            s.put("total_score", String.format("%.2f", 88.0 + i * 0.5));
            items.add(s);
        }
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
