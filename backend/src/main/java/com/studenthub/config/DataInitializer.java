package com.studenthub.config;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.studenthub.modules.sys.entity.SysRole;
import com.studenthub.modules.sys.entity.SysUser;
import com.studenthub.modules.sys.mapper.SysRoleMapper;
import com.studenthub.modules.sys.mapper.SysUserMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;

@Component
public class DataInitializer implements CommandLineRunner {

    private static final Logger log = LoggerFactory.getLogger(DataInitializer.class);

    private final SysUserMapper sysUserMapper;
    private final SysRoleMapper sysRoleMapper;

    public DataInitializer(SysUserMapper sysUserMapper, SysRoleMapper sysRoleMapper) {
        this.sysUserMapper = sysUserMapper;
        this.sysRoleMapper = sysRoleMapper;
    }

    @Override
    public void run(String... args) throws Exception {
        seedRoles();
        seedAdminUser();
    }

    private void seedRoles() {
        String[][] roles = {
                {"R-SY-ADMIN", "系统管理员", "school", "校级系统管理员，拥有所有模块权限"},
                {"R-SY-LEAGUE", "校团委管理员", "school", "校级团委管理员"},
                {"R-SY-AFFAIRS", "学生处管理员", "school", "学生处管理员"},
                {"R-COL-LEAGUE", "院系团委书记", "college", "院系级团委书记"},
                {"R-COL-COUN", "院系辅导员", "college", "院系级辅导员"},
                {"R-STU-NORM", "普通学生", "student", "普通学生"},
                {"R-STU-LEAGUE", "团支书", "student", "团支部书记"},
                {"R-STU-ASSOC", "社团社长/干部", "student", "社团干部"},
                {"R-STU-COMMUNITY", "楼层长/寝室长", "student", "社区自治干部"}
        };

        for (String[] r : roles) {
            Long count = sysRoleMapper.selectCount(new LambdaQueryWrapper<SysRole>().eq(SysRole::getCode, r[0]));
            if (count == 0) {
                SysRole role = new SysRole();
                role.setCode(r[0]);
                role.setName(r[1]);
                role.setScope(r[2]);
                role.setDescription(r[3]);
                role.setCreatedAt(LocalDateTime.now());
                role.setUpdatedAt(LocalDateTime.now());
                role.setIsDeleted(0);
                sysRoleMapper.insert(role);
                log.info("Seeded role: {}", r[0]);
            }
        }
    }

    private void seedAdminUser() {
        Long count = sysUserMapper.selectCount(new LambdaQueryWrapper<SysUser>().eq(SysUser::getUsername, "admin"));
        if (count == 0) {
            SysUser admin = new SysUser();
            admin.setUsername("admin");
            admin.setPasswordHash("admin@123");
            admin.setDisplayName("系统管理员");
            admin.setUserType("admin");
            admin.setStatus("active");
            admin.setCreatedAt(LocalDateTime.now());
            admin.setUpdatedAt(LocalDateTime.now());
            admin.setIsDeleted(0);
            sysUserMapper.insert(admin);
            log.info("Seeded admin user (admin / admin@123)");
        }
    }
}
