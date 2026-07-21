package com.studenthub;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.studenthub.common.R;
import com.studenthub.modules.idx.controller.IdxModuleController;
import com.studenthub.modules.qg.controller.QgModuleController;
import com.studenthub.modules.sq.controller.SqModuleController;
import com.studenthub.modules.st.controller.StModuleController;
import com.studenthub.modules.sys.entity.SysRole;
import com.studenthub.modules.sys.entity.SysUser;
import com.studenthub.modules.sys.mapper.SysRoleMapper;
import com.studenthub.modules.sys.mapper.SysUserMapper;
import com.studenthub.modules.ty.controller.TyModuleController;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;

@SpringBootTest
public class StudentHubCrudTest {

    @Autowired
    private SysUserMapper sysUserMapper;

    @Autowired
    private SysRoleMapper sysRoleMapper;

    @Autowired
    private IdxModuleController idxModuleController;

    @Autowired
    private TyModuleController tyModuleController;

    @Autowired
    private StModuleController stModuleController;

    @Autowired
    private SqModuleController sqModuleController;

    @Autowired
    private QgModuleController qgModuleController;

    @Test
    public void testUserCrud() {
        String testUsername = "test_user_" + System.currentTimeMillis();

        // 1. Create (Insert)
        SysUser testUser = new SysUser();
        testUser.setUsername(testUsername);
        testUser.setPasswordHash("password@123");
        testUser.setDisplayName("测试动态学生");
        testUser.setUserType("student");
        testUser.setStudentId(2026999999L);
        testUser.setStatus("active");
        testUser.setCreatedAt(LocalDateTime.now());
        testUser.setUpdatedAt(LocalDateTime.now());
        testUser.setIsDeleted(0);

        int insertResult = sysUserMapper.insert(testUser);
        Assertions.assertEquals(1, insertResult);
        Assertions.assertNotNull(testUser.getId());

        // 2. Read (Select)
        SysUser dbUser = sysUserMapper.selectOne(
                new LambdaQueryWrapper<SysUser>().eq(SysUser::getUsername, testUsername)
        );
        Assertions.assertNotNull(dbUser);
        Assertions.assertEquals("测试动态学生", dbUser.getDisplayName());

        // 3. Update
        dbUser.setDisplayName("测试动态学生(已改名)");
        dbUser.setUpdatedAt(LocalDateTime.now());
        int updateResult = sysUserMapper.updateById(dbUser);
        Assertions.assertEquals(1, updateResult);

        SysUser updatedDbUser = sysUserMapper.selectById(dbUser.getId());
        Assertions.assertEquals("测试动态学生(已改名)", updatedDbUser.getDisplayName());

        // 4. Delete
        int deleteResult = sysUserMapper.deleteById(updatedDbUser.getId());
        Assertions.assertEquals(1, deleteResult);

        SysUser deletedUser = sysUserMapper.selectById(updatedDbUser.getId());
        Assertions.assertNull(deletedUser, "Deleted user should be null after deletion");
    }

    @Test
    public void testRoleQuery() {
        List<SysRole> roles = sysRoleMapper.selectList(
                new LambdaQueryWrapper<SysRole>().eq(SysRole::getIsDeleted, 0)
        );
        Assertions.assertFalse(roles.isEmpty(), "Roles table should be seeded");
    }

    @Test
    public void testAllModulesCrudApi() {
        // 1. IDX 学生模块测试
        R<Map<String, Object>> students = idxModuleController.getStudents(1, 20, null, null, null);
        Assertions.assertEquals(0, students.getCode());
        Assertions.assertNotNull(students.getData().get("items"));

        // 2. TY 团员发展模块测试（包含推优、培养、政审、发展大会、转正、花名册）
        R<Map<String, Object>> tyApps = tyModuleController.getApplications(1, 20, null, null, null);
        Assertions.assertEquals(0, tyApps.getCode());

        R<Map<String, Object>> tyRecs = tyModuleController.getRecommendationMeetings(1, 20);
        Assertions.assertEquals(0, tyRecs.getCode());

        R<Map<String, Object>> tyCults = tyModuleController.getCultivationRecords(1, 20);
        Assertions.assertEquals(0, tyCults.getCode());

        R<Map<String, Object>> tyPols = tyModuleController.getPoliticalReviews(1, 20, null);
        Assertions.assertEquals(0, tyPols.getCode());

        R<Map<String, Object>> tyMeets = tyModuleController.getDevelopmentMeetings(1, 20);
        Assertions.assertEquals(0, tyMeets.getCode());

        R<Map<String, Object>> tyProbs = tyModuleController.getProbationaryRecords(1, 20);
        Assertions.assertEquals(0, tyProbs.getCode());

        R<Map<String, Object>> tyMebs = tyModuleController.getMembers(1, 20, null);
        Assertions.assertEquals(0, tyMebs.getCode());

        // 3. ST 社团活动模块测试
        R<Map<String, Object>> stAssocs = stModuleController.getAssociations(1, 20, null, null);
        Assertions.assertEquals(0, stAssocs.getCode());

        R<Map<String, Object>> stActs = stModuleController.getActivities(1, 20, null, null);
        Assertions.assertEquals(0, stActs.getCode());

        R<Map<String, Object>> stRecs = stModuleController.getRecruitPlans(1, 20, null, null);
        Assertions.assertEquals(0, stRecs.getCode());

        R<Map<String, Object>> stApplies = stModuleController.getRecruitApplies(1, 20, null);
        Assertions.assertEquals(0, stApplies.getCode());

        // 4. SQ 学生社区模块测试
        R<Map<String, Object>> sqBuilds = sqModuleController.getBuildings(1, 20);
        Assertions.assertEquals(0, sqBuilds.getCode());

        R<List<Map<String, Object>>> sqTree = sqModuleController.getBuildingTree();
        Assertions.assertEquals(0, sqTree.getCode());

        R<Map<String, Object>> sqIncs = sqModuleController.getIncidents(1, 20, null, null, null);
        Assertions.assertEquals(0, sqIncs.getCode());

        R<Map<String, Object>> sqInsps = sqModuleController.getInspections(1, 20, null, null);
        Assertions.assertEquals(0, sqInsps.getCode());

        // 5. QG 勤工助学模块测试
        R<Map<String, Object>> qgPos = qgModuleController.getPositions(1, 20, null, null);
        Assertions.assertEquals(0, qgPos.getCode());

        R<Map<String, Object>> qgDiffs = qgModuleController.getDifficultyCerts(1, 20, null, null);
        Assertions.assertEquals(0, qgDiffs.getCode());

        R<Map<String, Object>> qgAtts = qgModuleController.getAttendances(1, 20, null);
        Assertions.assertEquals(0, qgAtts.getCode());

        R<Map<String, Object>> qgApplies = qgModuleController.getApplies(1, 20, null);
        Assertions.assertEquals(0, qgApplies.getCode());
    }
}
