package com.studenthub.modules.st.controller;

import com.studenthub.common.R;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;

@RestController
@RequestMapping("/st")
public class StModuleController {

    private final JdbcTemplate jdbcTemplate;

    public StModuleController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

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

    @GetMapping("/associations")
    public R<Map<String, Object>> getAssociations() {
        List<Map<String, Object>> items = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> a = new HashMap<>();
            a.put("id", (long) i);
            a.put("assoc_code", String.format("ST-%02d", i));
            a.put("name", ASSOC_NAMES[i - 1]);
            a.put("category", i % 2 == 0 ? "学术科技类" : "文化体育类");
            a.put("president_name", NAMES[i - 1]);
            a.put("tutor_name", "王辅导员");
            a.put("star_level", 3 + (i % 3));
            a.put("star_rating", 3 + (i % 3));
            a.put("status", "active");
            a.put("member_count", 50 + i * 5);
            a.put("created_at", String.format("2026-03-%02d 10:00:00", (i % 25) + 1));
            items.add(a);
        }
        return R.ok(wrapPage(items));
    }

    @GetMapping("/users")
    public R<List<Map<String, Object>>> getUsers() {
        List<Map<String, Object>> users = new ArrayList<>();
        for (int i = 1; i <= 5; i++) {
            Map<String, Object> u = new HashMap<>();
            u.put("id", (long) i);
            u.put("display_name", "张教师" + i);
            u.put("username", "teacher" + i);
            users.add(u);
        }
        return R.ok(users);
    }

    @GetMapping("/students")
    public R<List<Map<String, Object>>> getStudents() {
        List<Map<String, Object>> list = new ArrayList<>();
        for (int i = 1; i <= 20; i++) {
            Map<String, Object> s = new HashMap<>();
            s.put("id", (long) i);
            s.put("name", NAMES[i - 1]);
            s.put("student_no", String.format("20230101%02d", i));
            list.add(s);
        }
        return R.ok(list);
    }

    @GetMapping("/associations/{id}/founders")
    public R<List<Map<String, Object>>> getFounders(@PathVariable Long id) {
        List<Map<String, Object>> founders = new ArrayList<>();
        for (int i = 1; i <= 3; i++) {
            Map<String, Object> f = new HashMap<>();
            f.put("id", (long) i);
            f.put("student_name", NAMES[i - 1]);
            f.put("student_no", String.format("20230101%02d", i));
            f.put("role", "发起人");
            founders.add(f);
        }
        return R.ok(founders);
    }

    @GetMapping("/associations/{id}/members")
    public R<List<Map<String, Object>>> getMembers(@PathVariable Long id) {
        List<Map<String, Object>> members = new ArrayList<>();
        for (int i = 1; i <= 10; i++) {
            Map<String, Object> m = new HashMap<>();
            m.put("id", (long) i);
            m.put("student_name", NAMES[i - 1]);
            m.put("student_no", String.format("20230101%02d", i));
            m.put("role", i == 1 ? "社长" : "社员");
            members.add(m);
        }
        return R.ok(members);
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
            r.put("accepted_count", 10 + i);
            r.put("status", "S3");
            r.put("created_at", String.format("2026-03-%02d 11:30:00", (i % 25) + 1));
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
            app.put("created_at", String.format("2026-03-%02d 14:20:00", (i % 25) + 1));
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
            act.put("biz_no", String.format("ST-ACT-2026-%04d", i));
            act.put("title", String.format("第%d届%s年度主题活动", i, ASSOC_NAMES[i - 1]));
            act.put("association_name", ASSOC_NAMES[i - 1]);
            act.put("activity_date", String.format("2026-04-%02d", (i % 25) + 1));
            act.put("location", String.format("学生活动中心%d室", 100 + i));
            act.put("status", "S3");
            act.put("level", "B");
            act.put("budget_cents", 50000);
            act.put("created_at", String.format("2026-03-%02d 16:00:00", (i % 25) + 1));
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
