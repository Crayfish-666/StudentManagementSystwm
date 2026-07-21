package com.studenthub.config;

import cn.dev33.satoken.stp.StpInterface;
import com.studenthub.modules.sys.mapper.SysUserRoleMapper;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

/**
 * Sa-Token 权限/角色接口实现。
 * 从 sys_user_role + sys_role 表查询真实角色，不再硬编码。
 */
@Component
public class StpInterfaceImpl implements StpInterface {

    private final SysUserRoleMapper sysUserRoleMapper;

    public StpInterfaceImpl(SysUserRoleMapper sysUserRoleMapper) {
        this.sysUserRoleMapper = sysUserRoleMapper;
    }

    @Override
    public List<String> getPermissionList(Object loginId, String loginType) {
        // V1 版本采用角色级控制，不细粒度到权限点；返回角色 code 作为权限标识
        return getRoleList(loginId, loginType);
    }

    @Override
    public List<String> getRoleList(Object loginId, String loginType) {
        if (loginId == null) {
            return new ArrayList<>();
        }
        try {
            Long userId = Long.parseLong(String.valueOf(loginId));
            List<String> roles = sysUserRoleMapper.selectRoleCodesByUserId(userId);
            return roles != null ? roles : new ArrayList<>();
        } catch (NumberFormatException e) {
            return new ArrayList<>();
        }
    }
}
