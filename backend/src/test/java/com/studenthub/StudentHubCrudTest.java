package com.studenthub;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.studenthub.modules.sys.entity.SysRole;
import com.studenthub.modules.sys.entity.SysUser;
import com.studenthub.modules.sys.mapper.SysRoleMapper;
import com.studenthub.modules.sys.mapper.SysUserMapper;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

import java.time.LocalDateTime;
import java.util.List;

@SpringBootTest
public class StudentHubCrudTest {

    @Autowired
    private SysUserMapper sysUserMapper;

    @Autowired
    private SysRoleMapper sysRoleMapper;

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
}
