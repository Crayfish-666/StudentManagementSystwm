package com.studenthub.common;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.io.Serializable;
import java.util.UUID;

/**
 * 统一响应封包（符合 SRD §2.3.1）。
 * 字段：code / message / data / biz_code / request_id
 */
public class R<T> implements Serializable {
    private static final long serialVersionUID = 1L;

    private int code;
    private String message;
    private T data;
    @JsonProperty("biz_code")
    private String bizCode;
    @JsonProperty("request_id")
    private String requestId;

    public R() {}

    public R(int code, String message, T data, String bizCode, String requestId) {
        this.code = code;
        this.message = message;
        this.data = data;
        this.bizCode = bizCode;
        this.requestId = requestId;
    }

    public static <T> R<T> ok(T data) {
        R<T> r = new R<>();
        r.setCode(0);
        r.setMessage("ok");
        r.setData(data);
        r.setRequestId(generateRequestId());
        return r;
    }

    public static <T> R<T> ok() {
        return ok(null);
    }

    public static <T> R<T> fail(int code, String message) {
        R<T> r = new R<>();
        r.setCode(code);
        r.setMessage(message);
        r.setRequestId(generateRequestId());
        return r;
    }

    public static <T> R<T> fail(int code, String message, String bizCode) {
        R<T> r = new R<>();
        r.setCode(code);
        r.setMessage(message);
        r.setBizCode(bizCode);
        r.setRequestId(generateRequestId());
        return r;
    }

    public static <T> R<T> fail(String message) {
        return fail(1001, message);
    }

    private static String generateRequestId() {
        return UUID.randomUUID().toString().replace("-", "");
    }

    public int getCode() { return code; }
    public void setCode(int code) { this.code = code; }

    public String getMessage() { return message; }
    public void setMessage(String message) { this.message = message; }

    public T getData() { return data; }
    public void setData(T data) { this.data = data; }

    public String getBizCode() { return bizCode; }
    public void setBizCode(String bizCode) { this.bizCode = bizCode; }

    public String getRequestId() { return requestId; }
    public void setRequestId(String requestId) { this.requestId = requestId; }
}
