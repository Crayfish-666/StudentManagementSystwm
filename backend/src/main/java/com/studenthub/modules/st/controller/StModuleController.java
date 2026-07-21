package com.studenthub.modules.st.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/st")
public class StModuleController {

    private static final String[] NAMES = {
            "张伟", "王芳", "李娜", "刘洋", "陈杰",
            "杨光", "黄磊", "周敏", "吴强", "徐霞",
            "孙浩", "胡婷", "朱勇", "高丽", "林涛",
            "何静", "郭平", "马明", "罗军", "梁晨"
    };

    private static final String[] ASSOC_NAMES = {
            "计算机算法与编程社", "英语角交际协会", "吉他与流行音乐社", "轮滑与极限运动社", "汉服文化研究社",
            "机器人与AI创新社", "青年志愿者协会", "羽毛球羽健社", "摄影与视觉艺术社", "辩论与演讲社",
            "动漫与二次元同好会", "跆拳道协会", "心理健康互助社", "创客与3D打印社", "电影鉴赏协会",
            "乒乓球同好会", "书法与国画社", "电子竞技社", "数学建模协会", "合唱团"
    };

    private static final String[] CATEGORIES = {
            "学术科技类", "文化体育类", "志愿公益类", "艺术兴趣类", "创新创业类"
    };

    @GetMapping("/associations")
    public R<Map<String, Object>> getAssociations() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> a = new HashMap<>();
            a.put("id", (long) i);
            a.put("name", ASSOC_NAMES[i - 1]);
            a.put("category", CATEGORIES[(i - 1) % CATEGORIES.length]);
            a.put("president_name", NAMES[i - 1]);
            a.put("member_count", 50 + i * 5);
            a.put("star_level", 3 + (i % 3));
            items.add(a);
        }
        return R.ok(wrapPage(items));
    }

    @GetMapping("/recruit-plans")
    public R<Map<String, Object>> getRecruitPlans() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> r = new HashMap<>();
            r.put("id", (long) i);
            r.put("title", String.format("2026春季%s招新计划", ASSOC_NAMES[i - 1]));
            r.put("association_name", ASSOC_NAMES[i - 1]);
            r.put("target_count", 30 + i);
            r.put("applied_count", 15 + i);
            r.put("status", "recruiting");
            items.add(r);
        }
        return R.ok(wrapPage(items));
    }

    @GetMapping("/recruit-applies")
    public R<Map<String, Object>> getRecruitApplies() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> app = new HashMap<>();
            app.put("id", (long) i);
            app.put("student_name", NAMES[i - 1]);
            app.put("student_no", String.format("20230101%02d", i));
            app.put("association_name", ASSOC_NAMES[(i - 1) % ASSOC_NAMES.length]);
            app.put("status", i % 2 == 0 ? "approved" : "pending");
            items.add(app);
        }
        return R.ok(wrapPage(items));
    }

    @GetMapping("/activities")
    public R<Map<String, Object>> getActivities() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> act = new HashMap<>();
            act.put("id", (long) i);
            act.put("title", String.format("第%d届%s年度主题活动", i, ASSOC_NAMES[i - 1]));
            act.put("association_name", ASSOC_NAMES[i - 1]);
            act.put("activity_date", String.format("2026-04-%02d", (i % 25) + 1));
            act.put("location", String.format("学生活动中心%d室", 100 + i));
            act.put("status", "approved");
            items.add(act);
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
