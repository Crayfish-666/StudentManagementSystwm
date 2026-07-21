package com.studenthub.modules.idx.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/idx")
public class IdxModuleController {

    private final JdbcTemplate jdbcTemplate;

    public IdxModuleController(JdbcTemplate jdbcTemplate) {
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

    private static final String[] MAJORS = {
            "软件工程", "电子商务", "环境设计", "计算机科学与技术", "通信工程"
    };

    @GetMapping("/students")
    public R<Map<String, Object>> getStudents() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> s = new HashMap<>();
            s.put("id", (long) i);
            s.put("student_no", String.format("20230101%02d", i));
            s.put("name", NAMES[i - 1]);
            s.put("student_name", NAMES[i - 1]);
            s.put("gender", i % 2 == 1 ? "男" : "女");
            s.put("college_name", COLLEGES[(i - 1) % COLLEGES.length]);
            s.put("major_name", MAJORS[(i - 1) % MAJORS.length]);
            s.put("class_name", String.format("%s230%d班", MAJORS[(i - 1) % MAJORS.length].substring(0, 2), (i % 3) + 1));
            s.put("status", "在读");
            s.put("created_at", String.format("2026-03-%02d 08:00:00", (i % 25) + 1));
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
