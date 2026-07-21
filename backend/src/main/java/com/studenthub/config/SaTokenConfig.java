package com.studenthub.config;

import cn.dev33.satoken.interceptor.SaInterceptor;
import cn.dev33.satoken.router.SaRouter;
import cn.dev33.satoken.stp.StpUtil;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

/**
 * Sa-Token 拦截器配置。
 * 注意：application.yml 已配置 context-path=/api/v1，
 * Sa-Token 的 SaRouter.match 匹配的是 context-path 之后的路径，
 * 因此用 /** 而非 /api/v1/**。
 */
@Configuration
public class SaTokenConfig implements WebMvcConfigurer {

    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        registry.addInterceptor(new SaInterceptor(handle -> {
            SaRouter.match("/**")
                    .notMatch("/auth/login", "/auth/refresh", "/healthz", "/actuator/**")
                    .check(r -> StpUtil.checkLogin());
        })).addPathPatterns("/**");
    }
}
