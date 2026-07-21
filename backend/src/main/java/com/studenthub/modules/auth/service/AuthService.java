package com.studenthub.modules.auth.service;

import cn.dev33.satoken.stp.StpUtil;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.studenthub.modules.auth.dto.LoginRequest;
import com.studenthub.modules.auth.dto.LoginResponse;
import com.studenthub.modules.sys.entity.SysUser;
import com.studenthub.modules.sys.mapper.SysUserMapper;
import com.studenthub.modules.sys.mapper.SysUserRoleMapper;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Service
public class AuthService {

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

        if (!passwordEncoder.matches(request.getPassword(), user.getPasswordHash())) {
            throw new IllegalArgumentException("用户名或密码错误");
        }

        if (!"active".equalsIgnoreCase(user.getStatus())) {
            throw new IllegalArgumentException("该账户已被锁定或禁用");
        }

        StpUtil.login(user.getId());
        user.setLastLoginAt(LocalDateTime.now());
        sysUserMapper.updateById(user);

        List<String> roles = sysUserRoleMapper.selectRoleCodesByUserId(user.getId());

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
        List<String> roles = sysUserRoleMapper.selectRoleCodesByUserId(user.getId());
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
}
