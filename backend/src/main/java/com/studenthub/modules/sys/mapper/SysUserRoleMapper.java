package com.studenthub.modules.sys.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.studenthub.modules.sys.entity.SysUserRole;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface SysUserRoleMapper extends BaseMapper<SysUserRole> {

    /**
     * 查询用户的所有角色 code 列表。
     */
    @Select("SELECT r.code FROM sys_role r " +
            "INNER JOIN sys_user_role ur ON ur.role_id = r.id " +
            "WHERE ur.user_id = #{userId} AND r.is_deleted = 0")
    List<String> selectRoleCodesByUserId(@Param("userId") Long userId);
}
