package com.studenthub.modules.sys.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/sys")
public class MenuController {

    @GetMapping({"/menu", "/menus/mine"})
    public R<Map<String, Object>> getMyMenus() {
        List<Map<String, Object>> menus = new ArrayList<>();

        // 1. 工作台
        Map<String, Object> dashboard = new HashMap<>();
        dashboard.put("code", "dashboard");
        dashboard.put("title", "工作台");
        dashboard.put("icon", "Odometer");
        dashboard.put("path", "/dashboard");

        List<Map<String, Object>> dashboardChildren = new ArrayList<>();
        dashboardChildren.add(createSubMenu("cmp-dashboard", "管理驾驶舱", "/dashboard", "views/Dashboard.vue"));
        dashboardChildren.add(createSubMenu("cmp-ranking", "综合分排行", "/cmp/ranking", "views/cmp/ScoreRanking.vue"));
        dashboard.put("children", dashboardChildren);
        menus.add(dashboard);

        // 2. 团员发展
        Map<String, Object> ty = new HashMap<>();
        ty.put("code", "ty");
        ty.put("title", "团员发展");
        ty.put("icon", "Flag");
        ty.put("path", "/ty");

        List<Map<String, Object>> tyChildren = new ArrayList<>();
        tyChildren.add(createSubMenu("ty-application", "入团申请", "/ty/application", "views/ty/ApplicationList.vue"));
        tyChildren.add(createSubMenu("ty-approval", "审批中心", "/ty/approval", "views/ty/ApprovalCenter.vue"));
        tyChildren.add(createSubMenu("ty-recommendation-meeting", "支部推优大会", "/ty/recommendation-meeting", "views/ty/RecommendationMeetingList.vue"));
        tyChildren.add(createSubMenu("ty-cultivation", "培养记录管理", "/ty/cultivation", "views/ty/CultivationView.vue"));
        tyChildren.add(createSubMenu("ty-development-object", "发展对象管理", "/ty/development-object", "views/ty/DevelopmentObjectView.vue"));
        tyChildren.add(createSubMenu("ty-political-review", "政治审查管理", "/ty/political-review", "views/ty/PoliticalReviewView.vue"));
        tyChildren.add(createSubMenu("ty-development-meeting", "接收发展大会", "/ty/development-meeting", "views/ty/DevelopmentMeetingView.vue"));
        tyChildren.add(createSubMenu("ty-probationary", "预备团员转正", "/ty/probationary", "views/ty/ProbationaryView.vue"));
        tyChildren.add(createSubMenu("ty-member-roster", "团员花名册", "/ty/member-roster", "views/ty/MemberRoster.vue"));
        ty.put("children", tyChildren);
        menus.add(ty);

        // 3. 社团活动
        Map<String, Object> st = new HashMap<>();
        st.put("code", "st");
        st.put("title", "社团活动");
        st.put("icon", "Trophy");
        st.put("path", "/st");

        List<Map<String, Object>> stChildren = new ArrayList<>();
        stChildren.add(createSubMenu("st-association", "社团管理", "/st/association", "views/st/AssociationList.vue"));
        stChildren.add(createSubMenu("st-recruit-plan", "招新计划管理", "/st/recruit-plan", "views/st/RecruitPlanList.vue"));
        stChildren.add(createSubMenu("st-recruit-apply", "招新申请广场", "/st/recruit-apply", "views/st/RecruitApplyList.vue"));
        stChildren.add(createSubMenu("st-activity", "活动管理与审批", "/st/activity", "views/st/ActivityList.vue"));
        st.put("children", stChildren);
        menus.add(st);

        // 4. 学生社区
        Map<String, Object> sq = new HashMap<>();
        sq.put("code", "sq");
        sq.put("title", "学生社区");
        sq.put("icon", "House");
        sq.put("path", "/sq");

        List<Map<String, Object>> sqChildren = new ArrayList<>();
        sqChildren.add(createSubMenu("sq-building", "楼栋与寝室网格", "/sq/building", "views/sq/BuildingTree.vue"));
        sqChildren.add(createSubMenu("sq-inspection", "巡查记录大厅", "/sq/inspection", "views/sq/InspectionList.vue"));
        sqChildren.add(createSubMenu("sq-incident", "异常事件处置", "/sq/incident", "views/sq/IncidentList.vue"));
        sq.put("children", sqChildren);
        menus.add(sq);

        // 5. 勤工助学
        Map<String, Object> qg = new HashMap<>();
        qg.put("code", "qg");
        qg.put("title", "勤工助学");
        qg.put("icon", "Briefcase");
        qg.put("path", "/qg");

        List<Map<String, Object>> qgChildren = new ArrayList<>();
        qgChildren.add(createSubMenu("qg-difficulty", "困难认定库", "/qg/difficulty", "views/qg/DifficultyList.vue"));
        qgChildren.add(createSubMenu("qg-position", "岗位管理", "/qg/position", "views/qg/PositionList.vue"));
        qgChildren.add(createSubMenu("qg-attendance", "工时打卡与考勤", "/qg/attendance", "views/qg/AttendanceRecord.vue"));
        qg.put("children", qgChildren);
        menus.add(qg);

        // 6. 我的申请
        Map<String, Object> mine = new HashMap<>();
        mine.put("code", "mine");
        mine.put("title", "我的申请");
        mine.put("icon", "Document");
        mine.put("path", "/mine");

        List<Map<String, Object>> mineChildren = new ArrayList<>();
        mineChildren.add(createSubMenu("mine-ty-development", "我的团员发展", "/mine/ty-development", "views/ty/MyDevelopment.vue"));
        mineChildren.add(createSubMenu("mine-ty-application", "我的入团申请", "/mine/ty-application", "views/ty/ApplicationList.vue"));
        mineChildren.add(createSubMenu("mine-thought-report", "我的思想汇报", "/mine/thought-report", "views/ty/MyThoughtReport.vue"));
        mineChildren.add(createSubMenu("mine-activity", "我的社团履历", "/mine/activity", "views/st/ActivityList.vue"));
        mineChildren.add(createSubMenu("mine-work", "我的勤工记录", "/mine/work", "views/qg/AttendanceRecord.vue"));
        mineChildren.add(createSubMenu("mine-score", "我的综合分", "/mine/score", "views/cmp/MyScore.vue"));
        mineChildren.add(createSubMenu("mine-profile", "我的学籍档案", "/mine/profile", "views/idx/MyProfile.vue"));
        mine.put("children", mineChildren);
        menus.add(mine);

        // 7. 学生管理
        Map<String, Object> idx = new HashMap<>();
        idx.put("code", "idx");
        idx.put("title", "学生管理");
        idx.put("icon", "User");
        idx.put("path", "/idx");

        List<Map<String, Object>> idxChildren = new ArrayList<>();
        idxChildren.add(createSubMenu("idx-student", "学生列表与履历", "/idx/student", "views/idx/StudentList.vue"));
        idxChildren.add(createSubMenu("idx-import", "学生批量导入", "/idx/import", "views/idx/StudentImport.vue"));
        idx.put("children", idxChildren);
        menus.add(idx);

        // 8. 系统管理
        Map<String, Object> sys = new HashMap<>();
        sys.put("code", "sys");
        sys.put("title", "系统管理");
        sys.put("icon", "Setting");
        sys.put("path", "/sys");

        List<Map<String, Object>> sysChildren = new ArrayList<>();
        sysChildren.add(createSubMenu("sys-dict", "字典管理", "/sys/dict", "views/sys/DictManage.vue"));
        sysChildren.add(createSubMenu("sys-user", "用户账号管理", "/sys/user", "views/sys/UserManage.vue"));
        sysChildren.add(createSubMenu("sys-org", "组织机构树", "/sys/org", "views/sys/OrgManage.vue"));
        sysChildren.add(createSubMenu("sys-job", "定时任务监控", "/sys/job", "views/sys/JobMonitor.vue"));
        sys.put("children", sysChildren);
        menus.add(sys);

        Map<String, Object> result = new HashMap<>();
        result.put("menus", menus);
        return R.ok(result);
    }

    private Map<String, Object> createSubMenu(String code, String title, String path, String component) {
        Map<String, Object> item = new HashMap<>();
        item.put("code", code);
        item.put("title", title);
        item.put("path", path);
        item.put("component", component);
        return item;
    }
}
