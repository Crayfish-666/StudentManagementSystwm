package com.studenthub.modules.auth.dto;

import java.io.Serializable;
import java.util.List;

public class LoginResponse implements Serializable {
    private static final long serialVersionUID = 1L;

    private String token;
    private String tokenName;
    private Long userId;
    private String username;
    private String displayName;
    private String userType;
    private Long studentId;
    private List<String> roles;

    public LoginResponse() {}

    public LoginResponse(String token, String tokenName, Long userId, String username, String displayName, String userType, Long studentId, List<String> roles) {
        this.token = token;
        this.tokenName = tokenName;
        this.userId = userId;
        this.username = username;
        this.displayName = displayName;
        this.userType = userType;
        this.studentId = studentId;
        this.roles = roles;
    }

    public String getToken() { return token; }
    public void setToken(String token) { this.token = token; }

    public String getTokenName() { return tokenName; }
    public void setTokenName(String tokenName) { this.tokenName = tokenName; }

    public Long getUserId() { return userId; }
    public void setUserId(Long userId) { this.userId = userId; }

    public String getUsername() { return username; }
    public void setUsername(String username) { this.username = username; }

    public String getDisplayName() { return displayName; }
    public void setDisplayName(String displayName) { this.displayName = displayName; }

    public String getUserType() { return userType; }
    public void setUserType(String userType) { this.userType = userType; }

    public Long getStudentId() { return studentId; }
    public void setStudentId(Long studentId) { this.studentId = studentId; }

    public List<String> getRoles() { return roles; }
    public void setRoles(List<String> roles) { this.roles = roles; }
}
