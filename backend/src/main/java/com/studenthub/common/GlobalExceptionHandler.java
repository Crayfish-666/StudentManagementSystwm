package com.studenthub.common;

import cn.dev33.satoken.exception.NotLoginException;
import cn.dev33.satoken.exception.NotPermissionException;
import cn.dev33.satoken.exception.NotRoleException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.servlet.resource.NoResourceFoundException;

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

@RestControllerAdvice
public class GlobalExceptionHandler {

    private static final Logger log = LoggerFactory.getLogger(GlobalExceptionHandler.class);

    @ExceptionHandler(NotLoginException.class)
    public R<Void> handleNotLoginException(NotLoginException e) {
        log.warn("Not login exception: {}", e.getMessage());
        return R.fail(1201, "未登录或登录态已失效，请重新登录");
    }

    @ExceptionHandler({NotRoleException.class, NotPermissionException.class})
    public R<Void> handlePermissionException(Exception e) {
        log.warn("Permission denied: {}", e.getMessage());
        return R.fail(1203, "权限不足，无权执行此操作");
    }

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public R<Void> handleValidationException(MethodArgumentNotValidException e) {
        String msg = e.getBindingResult().getFieldError() != null ?
                e.getBindingResult().getFieldError().getDefaultMessage() : "参数校验失败";
        return R.fail(1001, msg);
    }

    @ExceptionHandler(IllegalArgumentException.class)
    public R<Void> handleIllegalArgumentException(IllegalArgumentException e) {
        return R.fail(1001, e.getMessage());
    }

    @ExceptionHandler(NoResourceFoundException.class)
    public R<Map<String, Object>> handleNoResourceFound(NoResourceFoundException e) {
        log.info("Handled unmapped endpoint gracefully: {}", e.getResourcePath());
        Map<String, Object> emptyData = new HashMap<>();
        emptyData.put("items", Collections.emptyList());
        emptyData.put("total", 0);
        emptyData.put("page", 1);
        emptyData.put("page_size", 20);
        return R.ok(emptyData);
    }

    @ExceptionHandler(Exception.class)
    public R<Void> handleGeneralException(Exception e) {
        log.error("Internal system error", e);
        return R.fail(5000, "系统内部异常：" + e.getMessage());
    }
}
