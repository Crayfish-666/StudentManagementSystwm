package com.studenthub.modules.idx.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/idx")
public class IdxModuleController {

    @GetMapping("/students")
    public R<Map<String, Object>> getStudents() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> s1 = new HashMap<>();
        s1.put("id", 1L);
        s1.put("student_no", "2023010101");
        s1.put("name", "张三");
        s1.put("gender", "男");
        s1.put("college_name", "计算机学院");
        s1.put("major_name", "软件工程");
        s1.put("class_name", "软工2301班");
        s1.put("status", "在读");
        items.add(s1);

        Map<String, Object> s2 = new HashMap<>();
        s2.put("id", 2L);
        s2.put("student_no", "2023010102");
        s2.put("name", "李四");
        s2.put("gender", "女");
        s2.put("college_name", "经济管理学院");
        s2.put("major_name", "电子商务");
        s2.put("class_name", "电商2302班");
        s2.put("status", "在读");
        items.add(s2);
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
