package com.studenthub.modules.sq.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/sq")
public class SqModuleController {

    @GetMapping("/buildings")
    public R<Map<String, Object>> getBuildings() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> b = new HashMap<>();
        b.put("id", 1L);
        b.put("name", "学生公寓1号楼");
        b.put("code", "SQ-BUILD-01");
        b.put("room_count", 120);
        items.add(b);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/inspections")
    public R<Map<String, Object>> getInspections() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> ins = new HashMap<>();
        ins.put("id", 1L);
        ins.put("building_name", "学生公寓1号楼");
        ins.put("room_no", "302");
        ins.put("score", 95);
        ins.put("inspector_name", "王网格员");
        ins.put("inspect_date", "2026-03-18");
        items.add(ins);
        return R.ok(wrapPage(items));
    }

    @GetMapping("/incidents")
    public R<Map<String, Object>> getIncidents() {
        List<Map<String, Object>> items = new ArrayList<>();
        Map<String, Object> inc = new HashMap<>();
        inc.put("id", 1L);
        inc.put("building_name", "学生公寓2号楼");
        inc.put("room_no", "405");
        inc.put("type", "电器违规使用");
        inc.put("status", "resolved");
        items.add(inc);
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
