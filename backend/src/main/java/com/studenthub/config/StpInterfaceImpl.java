package com.studenthub.config;

import cn.dev33.satoken.stp.StpInterface;
import org.springframework.stereotype.Component;
import java.util.ArrayList;
import java.util.List;

@Component
public class StpInterfaceImpl implements StpInterface {

    @Override
    public List<String> getPermissionList(Object loginId, String loginType) {
        List<String> permissions = new ArrayList<>();
        permissions.add("*"); // 开发调试阶段给予全量权限
        return permissions;
    }

    @Override
    public List<String> getRoleList(Object loginId, String loginType) {
        List<String> roles = new ArrayList<>();
        if ("1".equals(String.valueOf(loginId))) {
            roles.add("R-SY-ADMIN");
        } else {
            roles.add("R-STU-NORM");
        }
        return roles;
    }
}
