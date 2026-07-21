package com.studenthub.config;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.studenthub.modules.sys.entity.SysRole;
import com.studenthub.modules.sys.entity.SysUser;
import com.studenthub.modules.sys.entity.SysUserRole;
import com.studenthub.modules.sys.mapper.SysRoleMapper;
import com.studenthub.modules.sys.mapper.SysUserMapper;
import com.studenthub.modules.sys.mapper.SysUserRoleMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.CommandLineRunner;
import org.springframework.core.annotation.Order;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;

/**
 * 数据初始化器。
 * 使用 BCrypt 加密密码（cost=12，符合 ADR-005）。
 * 兼容旧数据：如果已有用户密码是明文，自动升级为 BCrypt。
 * 同时 seed sys_user_role 关联表，确保 RBAC 数据基础完整。
 */
@Component
@Order(1)
public class DataInitializer implements CommandLineRunner {

    private static final Logger log = LoggerFactory.getLogger(DataInitializer.class);

    private final SysUserMapper sysUserMapper;
    private final SysRoleMapper sysRoleMapper;
    private final SysUserRoleMapper sysUserRoleMapper;
    private final PasswordEncoder passwordEncoder;

    public DataInitializer(SysUserMapper sysUserMapper,
                           SysRoleMapper sysRoleMapper,
                           SysUserRoleMapper sysUserRoleMapper,
                           PasswordEncoder passwordEncoder) {
        this.sysUserMapper = sysUserMapper;
        this.sysRoleMapper = sysRoleMapper;
        this.sysUserRoleMapper = sysUserRoleMapper;
        this.passwordEncoder = passwordEncoder;
    }

    @Override
    public void run(String... args) {
        try {
            seedRoles();
            seedAdminUser();
            seedCounselorUser();
            seed20StudentUsers();
            seedUserRoles();
            upgradePlaintextPasswords();
            log.info("Data initialization completed successfully.");
        } catch (Exception e) {
            log.error("Data initialization failed, application continues running.", e);
        }
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
            try {
                Long count = sysRoleMapper.selectCount(new LambdaQueryWrapper<SysRole>().eq(SysRole::getCode, r[0]));
                if (count == null || count == 0) {
                    SysRole role = new SysRole();
                    role.setCode(r[0]);
                    role.setName(r[1]);
                    role.setScope(r[2]);
                    role.setDescription(r[3]);
                    role.setCreatedAt(LocalDateTime.now());
                    role.setUpdatedAt(LocalDateTime.now());
                    role.setIsDeleted(0);
                    sysRoleMapper.insert(role);
                }
            } catch (Exception e) {
                log.warn("Failed to seed role {}: {}", r[0], e.getMessage());
            }
        }
    }

    private void seedAdminUser() {
        seedUserIfAbsent("admin", "admin@123", "系统管理员", "admin", null);
    }

    private void seedCounselorUser() {
        seedUserIfAbsent("counselor", "counselor@123", "张辅导员", "counselor", null);
    }

    private void seed20StudentUsers() {
        String[] studentNames = {
                "张伟", "王芳", "李娜", "刘洋", "陈杰",
                "杨光", "黄磊", "周敏", "吴强", "徐霞",
                "孙浩", "胡婷", "朱勇", "高丽", "林涛",
                "何静", "郭平", "马明", "罗军", "梁晨"
        };

        for (int i = 0; i < studentNames.length; i++) {
            String studentNo = String.format("20230101%02d", i + 1);
            Long studentId = 2023010100L + i + 1;
            seedUserIfAbsent(studentNo, "student@123", studentNames[i], "student", studentId);
        }
        log.info("Verified 20 student accounts.");
    }

    private void seedUserIfAbsent(String username, String rawPassword, String displayName,
                                   String userType, Long studentId) {
        try {
            Long count = sysUserMapper.selectCount(new LambdaQueryWrapper<SysUser>().eq(SysUser::getUsername, username));
            if (count != null && count > 0) {
                return;
            }
            SysUser user = new SysUser();
            user.setUsername(username);
            user.setPasswordHash(passwordEncoder.encode(rawPassword));
            user.setDisplayName(displayName);
            user.setUserType(userType);
            user.setStudentId(studentId);
            user.setStatus("active");
            user.setCreatedAt(LocalDateTime.now());
            user.setUpdatedAt(LocalDateTime.now());
            user.setIsDeleted(0);
            sysUserMapper.insert(user);
        } catch (Exception e) {
            log.warn("Failed to seed user {}: {}", username, e.getMessage());
        }
    }

    private void seedUserRoles() {
        assignRole("admin", "R-SY-ADMIN");
        assignRole("counselor", "R-COL-COUN");
        for (int i = 1; i <= 20; i++) {
            String studentNo = String.format("20230101%02d", i);
            assignRole(studentNo, "R-STU-NORM");
        }
        log.info("Verified user-role associations.");
    }

    private void assignRole(String username, String roleCode) {
        try {
            SysUser user = sysUserMapper.selectOne(new LambdaQueryWrapper<SysUser>().eq(SysUser::getUsername, username));
            SysRole role = sysRoleMapper.selectOne(new LambdaQueryWrapper<SysRole>().eq(SysRole::getCode, roleCode));
            if (user == null || role == null) {
                return;
            }
            Long count = sysUserRoleMapper.selectCount(
                    new LambdaQueryWrapper<SysUserRole>()
                            .eq(SysUserRole::getUserId, user.getId())
                            .eq(SysUserRole::getRoleId, role.getId()));
            if (count != null && count > 0) {
                return;
            }
            SysUserRole ur = new SysUserRole();
            ur.setUserId(user.getId());
            ur.setRoleId(role.getId());
            ur.setCreatedAt(LocalDateTime.now());
            sysUserRoleMapper.insert(ur);
        } catch (Exception e) {
            log.warn("Failed to assign role {} to user {}: {}", roleCode, username, e.getMessage());
        }
    }

    /**
     * 升级旧明文密码为 BCrypt（首次启动时批量处理）。
     */
    private void upgradePlaintextPasswords() {
        try {
            // 只检查已知的 3 个种子账户，避免全表扫描
            upgradeIfPlaintext("admin", "admin@123");
            upgradeIfPlaintext("counselor", "counselor@123");
            for (int i = 1; i <= 20; i++) {
                String studentNo = String.format("20230101%02d", i);
                upgradeIfPlaintext(studentNo, "student@123");
            }
        } catch (Exception e) {
            log.warn("Password upgrade check failed: {}", e.getMessage());
        }
    }

    private void upgradeIfPlaintext(String username, String knownPassword) {
        SysUser user = sysUserMapper.selectOne(new LambdaQueryWrapper<SysUser>().eq(SysUser::getUsername, username));
        if (user == null || user.getPasswordHash() == null) return;
        String hash = user.getPasswordHash();
        boolean isBcrypt = hash.startsWith("$2a$") || hash.startsWith("$2b$") || hash.startsWith("$2y$");
        if (!isBcrypt && hash.equals(knownPassword)) {
            log.info("Upgrading plaintext password to BCrypt for user: {}", username);
            user.setPasswordHash(passwordEncoder.encode(knownPassword));
            sysUserMapper.updateById(user);
        }
    }
}
