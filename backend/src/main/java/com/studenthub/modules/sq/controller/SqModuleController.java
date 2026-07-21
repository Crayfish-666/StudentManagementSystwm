package com.studenthub.modules.sq.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/sq")
public class SqModuleController {

    private static final String[] NAMES = {
            "张伟", "王芳", "李娜", "刘洋", "陈杰",
            "杨光", "黄磊", "周敏", "吴强", "徐霞",
            "孙浩", "胡婷", "朱勇", "高丽", "林涛",
            "何静", "郭平", "马明", "罗军", "梁晨"
    };

    @GetMapping("/buildings")
    public R<Map<String, Object>> getBuildings() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> b = new HashMap<>();
            b.put("id", (long) i);
            b.put("name", String.format("学生公寓%d号楼", i));
            b.put("code", String.format("SQ-BUILD-%02d", i));
            b.put("room_count", 100 + i * 2);
            items.add(b);
        }
        return R.ok(wrapPage(items));
    }

    @GetMapping("/inspections")
    public R<Map<String, Object>> getInspections() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> ins = new HashMap<>();
            ins.put("id", (long) i);
            ins.put("building_name", String.format("学生公寓%d号楼", (i % 6) + 1));
            ins.put("room_no", String.format("%d0%d", (i % 5) + 1, (i % 8) + 1));
            ins.put("score", 90 + (i % 10));
            ins.put("inspector_name", NAMES[i - 1]);
            ins.put("inspect_date", String.format("2026-03-%02d", (i % 25) + 1));
            items.add(ins);
        }
        return R.ok(wrapPage(items));
    }

    @GetMapping("/incidents")
    public R<Map<String, Object>> getIncidents() {
        List<Map<String, Object>> items = new ArrayList<>();
        String[] types = { "大功率违规违章用电", "宿舍夜不归宿", "走廊堆放易燃杂物", "公共设施报修", "大声喧哗打扰作息" };
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> inc = new HashMap<>();
            inc.put("id", (long) i);
            inc.put("building_name", String.format("学生公寓%d号楼", (i % 6) + 1));
            inc.put("room_no", String.format("%d0%d", (i % 5) + 1, (i % 8) + 1));
            inc.put("type", types[(i - 1) % types.length]);
            inc.put("status", i % 3 == 0 ? "pending" : "resolved");
            items.add(inc);
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
