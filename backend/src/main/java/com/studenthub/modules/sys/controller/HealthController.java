package com.studenthub.modules.sys.controller;

import com.studenthub.common.R;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.HashMap;
import java.util.Map;

@RestController
public class HealthController {

    @GetMapping("/healthz")
    public R<Map<String, Object>> healthz() {
        Map<String, Object> status = new HashMap<>();
        status.put("status", "UP");
        status.put("version", "2.1.0-SpringBoot");
        status.put("timestamp", System.currentTimeMillis());
        return R.ok(status);
    }
}
