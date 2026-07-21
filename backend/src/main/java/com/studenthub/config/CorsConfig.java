package com.studenthub.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;
import org.springframework.web.filter.CorsFilter;

import java.util.Arrays;
import java.util.List;

/**
 * CORS 跨域白名单配置。
 * 通过环境变量 CORS_ORIGINS 配置允许的域名（逗号分隔），
 * 默认允许开发环境 localhost:5173。
 */
@Configuration
public class CorsConfig {

    @Value("${CORS_ORIGINS:http://localhost:5173,http://127.0.0.1:5173}")
    private String corsOrigins;

    @Bean
    public CorsFilter corsFilter() {
        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        CorsConfiguration config = new CorsConfiguration();
        config.setAllowCredentials(true);

        // 白名单模式：仅允许环境变量中配置的 Origin
        List<String> allowedOrigins = Arrays.asList(corsOrigins.split(","));
        allowedOrigins.forEach(origin -> config.addAllowedOrigin(origin.trim()));

        config.addAllowedHeader("*");
        config.addAllowedMethod("*");
        config.addExposedHeader("X-Request-Id");
        source.registerCorsConfiguration("/**", config);
        return new CorsFilter(source);
    }
}
