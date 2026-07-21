package com.studenthub.modules.auth.service;

import cn.dev33.satoken.stp.StpUtil;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.studenthub.modules.auth.dto.LoginRequest;
import com.studenthub.modules.auth.dto.LoginResponse;
import com.studenthub.modules.sys.entity.SysUser;
import com.studenthub.modules.sys.mapper.SysUserMapper;
import com.studenthub.modules.sys.mapper.SysUserRoleMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.Collections;
import java.util.List;

/**
 * 认证服务。
 * 密码兼容策略：优先 BCrypt 校验；如果存储的不是 BCrypt 哈希（如历史明文），
 * 且明文匹配，则自动升级为 BCrypt 哈希后再返回。
 */
@Service
public class AuthService {

    private static final Logger log = LoggerFactory.getLogger(AuthService.class);

    private final SysUserMapper sysUserMapper;
    private final SysUserRoleMapper sysUserRoleMapper;
    private final PasswordEncoder passwordEncoder;

    public AuthService(SysUserMapper sysUserMapper,
                       SysUserRoleMapper sysUserRoleMapper,
                       PasswordEncoder passwordEncoder) {
        this.sysUserMapper = sysUserMapper;
        this.sysUserRoleMapper = sysUserRoleMapper;
        this.passwordEncoder = passwordEncoder;
    }

    public LoginResponse login(LoginRequest request) {
        SysUser user = sysUserMapper.selectOne(
                new LambdaQueryWrapper<SysUser>()
                        .eq(SysUser::getUsername, request.getUsername())
        );

        if (user == null) {
            throw new IllegalArgumentException("用户名或密码错误");
        }

        if (!verifyPasswordAndUpgrade(request.getPassword(), user)) {
            throw new IllegalArgumentException("用户名或密码错误");
        }

        if (!"active".equalsIgnoreCase(user.getStatus())) {
            throw new IllegalArgumentException("该账户已被锁定或禁用");
        }

        StpUtil.login(user.getId());
        user.setLastLoginAt(LocalDateTime.now());
        sysUserMapper.updateById(user);

        List<String> roles = Collections.emptyList();
        try {
            roles = sysUserRoleMapper.selectRoleCodesByUserId(user.getId());
            if (roles == null) roles = Collections.emptyList();
        } catch (Exception e) {
            log.warn("Failed to query roles for user {}: {}", user.getUsername(), e.getMessage());
        }

        return new LoginResponse(
                StpUtil.getTokenValue(),
                StpUtil.getTokenName(),
                user.getId(),
                user.getUsername(),
                user.getDisplayName(),
                user.getUserType(),
                user.getStudentId(),
                roles
        );
    }

    public void logout() {
        if (StpUtil.isLogin()) {
            StpUtil.logout();
        }
    }

    /**
     * 获取当前登录用户信息（供 /auth/me 端点使用）。
     */
    public LoginResponse getCurrentUser() {
        if (!StpUtil.isLogin()) {
            throw new IllegalArgumentException("未登录");
        }
        Long userId = StpUtil.getLoginIdAsLong();
        SysUser user = sysUserMapper.selectById(userId);
        if (user == null) {
            throw new IllegalArgumentException("用户不存在");
        }
        List<String> roles = Collections.emptyList();
        try {
            roles = sysUserRoleMapper.selectRoleCodesByUserId(userId);
            if (roles == null) roles = Collections.emptyList();
        } catch (Exception e) {
            log.warn("Failed to query roles for user {}: {}", user.getUsername(), e.getMessage());
        }
        return new LoginResponse(
                StpUtil.getTokenValue(),
                StpUtil.getTokenName(),
                user.getId(),
                user.getUsername(),
                user.getDisplayName(),
                user.getUserType(),
                user.getStudentId(),
                roles
        );
    }

    /**
     * 验证密码，如果存储的不是 BCrypt 哈希且明文匹配，则自动升级为 BCrypt。
     */
    private boolean verifyPasswordAndUpgrade(String rawPassword, SysUser user) {
        String storedHash = user.getPasswordHash();
        if (storedHash == null || storedHash.isEmpty()) {
            return false;
        }

        // BCrypt 哈希以 $2a$ 或 $2b$ 开头
        boolean isBcrypt = storedHash.startsWith("$2a$") || storedHash.startsWith("$2b$") || storedHash.startsWith("$2y$");

        if (isBcrypt) {
            try {
                return passwordEncoder.matches(rawPassword, storedHash);
            } catch (Exception e) {
                log.warn("BCrypt check failed for user {}: {}", user.getUsername(), e.getMessage());
                return false;
            }
        }

        // 兼容旧明文密码：匹配则升级为 BCrypt
        if (storedHash.equals(rawPassword)) {
            log.info("Upgrading plaintext password to BCrypt for user: {}", user.getUsername());
            user.setPasswordHash(passwordEncoder.encode(rawPassword));
            sysUserMapper.updateById(user);
            return true;
        }

        return false;
    }
}
