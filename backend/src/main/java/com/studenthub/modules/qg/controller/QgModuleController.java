package com.studenthub.modules.qg.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/qg")
public class QgModuleController {

    private static final String[] NAMES = {
            "张伟", "王芳", "李娜", "刘洋", "陈杰",
            "杨光", "黄磊", "周敏", "吴强", "徐霞",
            "孙浩", "胡婷", "朱勇", "高丽", "林涛",
            "何静", "郭平", "马明", "罗军", "梁晨"
    };

    private static final String[] POSITIONS = {
            "图书馆图书整理助理", "教务处档案整理助理", "实验中心设备维护助理", "学工处日常事务助理", "体育馆场地管理助理",
            "网络中心技术运维助理", "后勤服务中心监督员", "心理咨询中心接待助理", "校团委活动策划助理", "各院系办公室助理"
    };

    @GetMapping("/difficulties")
    public R<Map<String, Object>> getDifficulties() {
        List<Map<String, Object>> items = new ArrayList<>();
        String[] levels = { "特别困难", "一般困难", "特殊困难" };
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> d = new HashMap<>();
            d.put("id", (long) i);
            d.put("student_name", NAMES[i - 1]);
            d.put("student_no", String.format("20230101%02d", i));
            d.put("level", levels[(i - 1) % levels.length]);
            d.put("status", "approved");
            items.add(d);
        }
        return R.ok(wrapPage(items));
    }

    @GetMapping("/positions")
    public R<Map<String, Object>> getPositions() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> p = new HashMap<>();
            p.put("id", (long) i);
            p.put("title", String.format("%s (岗位%d)", POSITIONS[(i - 1) % POSITIONS.length], i));
            p.put("department", "公共管理与服务部门");
            p.put("hourly_rate", 20.0 + (i % 5));
            p.put("quota", 5 + i);
            items.add(p);
        }
        return R.ok(wrapPage(items));
    }

    @GetMapping("/attendances")
    public R<Map<String, Object>> getAttendances() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> a = new HashMap<>();
            a.put("id", (long) i);
            a.put("student_name", NAMES[i - 1]);
            a.put("position_title", POSITIONS[(i - 1) % POSITIONS.length]);
            a.put("work_date", String.format("2026-03-%02d", (i % 25) + 1));
            a.put("hours", 2.0 + (i % 4));
            a.put("status", "approved");
            items.add(a);
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
