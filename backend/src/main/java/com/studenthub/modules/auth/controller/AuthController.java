package com.studenthub.modules.auth.controller;

import cn.dev33.satoken.stp.StpUtil;
import com.studenthub.common.R;
import com.studenthub.modules.auth.dto.LoginRequest;
import com.studenthub.modules.auth.dto.LoginResponse;
import com.studenthub.modules.auth.service.AuthService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/auth")
@RequiredArgsConstructor
public class AuthController {

    private final AuthService authService;

    @PostMapping("/login")
    public R<LoginResponse> login(@Valid @RequestBody LoginRequest request) {
        LoginResponse response = authService.login(request);
        return R.ok(response);
    }

    @PostMapping("/logout")
    public R<Void> logout() {
        authService.logout();
        return R.ok();
    }

    @GetMapping("/check")
    public R<Boolean> checkLogin() {
        return R.ok(StpUtil.isLogin());
    }
}
