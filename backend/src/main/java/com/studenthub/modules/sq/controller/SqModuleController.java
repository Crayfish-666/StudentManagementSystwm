package com.studenthub.modules.sq.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/sq")
public class SqModuleController {

    private final JdbcTemplate jdbcTemplate;

    public SqModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

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
            b.put("code", String.format("SQ-B%02d", i));
            b.put("name", String.format("学生公寓%d号楼", i));
            b.put("total_floors", 6);
            b.put("room_count", 120);
            b.put("tutor_name", NAMES[(i - 1) % NAMES.length] + "辅导员");
            b.put("created_at", String.format("2026-03-%02d 08:30:00", (i % 25) + 1));
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
            ins.put("created_at", String.format("2026-03-%02d 10:45:00", (i % 25) + 1));
            items.add(ins);
        }
        return R.ok(wrapPage(items));
    }

    @GetMapping("/incidents")
    public R<Map<String, Object>> getIncidents() {
        List<Map<String, Object>> items = new ArrayList<>();
        String[] types = { "大功率违规用电", "夜不归宿", "公共设施报修", "大声喧哗打扰作息", "私拉乱接电线" };
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> inc = new HashMap<>();
            inc.put("id", (long) i);
            inc.put("biz_no", String.format("SQ-INC-2026-%04d", i));
            inc.put("building_name", String.format("学生公寓%d号楼", (i % 6) + 1));
            inc.put("room_no", String.format("%d0%d", (i % 5) + 1, (i % 8) + 1));
            inc.put("type", types[(i - 1) % types.length]);
            inc.put("incident_type", types[(i - 1) % types.length]);
            inc.put("status", i % 3 == 0 ? "pending" : "resolved");
            inc.put("description", String.format("宿管员于%d号楼巡查时发现该寝室存在%s隐患，已登记处理。", (i % 6) + 1, types[(i - 1) % types.length]));
            inc.put("created_at", String.format("2026-03-%02d 21:15:00", (i % 25) + 1));
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
