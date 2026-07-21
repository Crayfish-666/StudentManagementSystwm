package com.studenthub.common;

import cn.dev33.satoken.exception.NotLoginException;
import cn.dev33.satoken.exception.NotPermissionException;
import cn.dev33.satoken.exception.NotRoleException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.servlet.resource.NoResourceFoundException;

/**
 * 全局异常处理（符合 SRD §2.3 错误码段）。
 * 错误码段：0 成功 / 1000-1099 通用 / 1100-1199 IDX / 2000-2099 TY /
 * 3000-3099 ST / 4000-4099 SQ / 5000-5099 QG / 6000-6099 CMP /
 * 7000-7099 NOTI / 8000-8099 FILE / 9000-9099 SYS
 */
@RestControllerAdvice
public class GlobalExceptionHandler {

    private static final Logger log = LoggerFactory.getLogger(GlobalExceptionHandler.class);

    @ExceptionHandler(NotLoginException.class)
    public R<Void> handleNotLoginException(NotLoginException e) {
        log.warn("Not login exception: {}", e.getMessage());
        return R.fail(10401, "未登录或登录态已失效，请重新登录", "AUTH.NOT_LOGIN");
    }

    @ExceptionHandler({NotRoleException.class, NotPermissionException.class})
    public R<Void> handlePermissionException(Exception e) {
        log.warn("Permission denied: {}", e.getMessage());
        return R.fail(10403, "权限不足，无权执行此操作", "AUTH.FORBIDDEN");
    }

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public R<Void> handleValidationException(MethodArgumentNotValidException e) {
        String msg = e.getBindingResult().getFieldError() != null ?
                e.getBindingResult().getFieldError().getDefaultMessage() : "参数校验失败";
        return R.fail(10422, msg, "COMMON.VALIDATION_FAILED");
    }

    @ExceptionHandler(IllegalArgumentException.class)
    public R<Void> handleIllegalArgumentException(IllegalArgumentException e) {
        return R.fail(1001, e.getMessage(), "COMMON.ILLEGAL_ARGUMENT");
    }

    /**
     * 未映射端点返回 HTTP 404（而非静默吞成 200+空列表，避免前端误以为成功）。
     */
    @ExceptionHandler(NoResourceFoundException.class)
    public ResponseEntity<R<Void>> handleNoResourceFound(NoResourceFoundException e) {
        log.warn("Resource not found: {}", e.getResourcePath());
        return ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(R.fail(10404, "请求的资源不存在: " + e.getResourcePath(), "COMMON.NOT_FOUND"));
    }

    @ExceptionHandler(Exception.class)
    public R<Void> handleGeneralException(Exception e) {
        log.error("Internal system error", e);
        return R.fail(1500, "系统内部异常", "COMMON.INTERNAL_ERROR");
    }
}
