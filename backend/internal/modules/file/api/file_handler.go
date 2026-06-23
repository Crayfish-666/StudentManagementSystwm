package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/file/service"
	"student-system/pkg/response"
)

// 允许的 MIME 类型前缀
var allowedMIMETypes = []string{
	"image/",
	"application/pdf",
	"application/msword",
	"application/vnd.openxmlformats-officedocument.",
	"text/",
}

// 最大文件大小 50MB
const maxFileSize = 50 << 20

// FileHandler 文件接口处理器。
type FileHandler struct {
	svc *service.FileService
}

// NewFileHandler 创建文件处理器。
func NewFileHandler(svc *service.FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

// Upload 上传文件。POST /api/v1/files/upload
func (h *FileHandler) Upload(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	// 限制文件大小
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, 40001, "请选择要上传的文件")
		return
	}
	defer file.Close()

	// 校验文件大小
	if header.Size > maxFileSize {
		response.Fail(c, 40001, "文件大小不能超过 50MB")
		return
	}

	// 校验 MIME 类型
	contentType := header.Header.Get("Content-Type")
	if !isAllowedMIME(contentType) {
		response.Fail(c, 40001, "不支持的文件类型")
		return
	}

	module := c.PostForm("module")
	bizType := c.PostForm("biz_type")
	if module == "" || bizType == "" {
		response.Fail(c, 40001, "module 和 biz_type 不能为空")
		return
	}

	originalName := header.Filename

	meta, err := h.svc.Upload(c.Request.Context(), userID, module, bizType, originalName, file, header.Size)
	if err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}

	response.OK(c, gin.H{
		"file_id": meta.ID,             // 文件元数据主键（供业务表 FK 引用）
		"id":      meta.ID,             // 兼容别名
		"key":     meta.StorageKey,
		"url":     "/api/v1/files/" + meta.StorageKey,
		"hash":    "sha256:" + meta.SHA256,
		"size":    meta.SizeBytes,
	})
}

// Download 下载文件。GET /api/v1/files/*key
func (h *FileHandler) Download(c *gin.Context) {
	// 通配符捕获的 key 会以 "/" 开头，去掉前缀
	key := c.Param("key")
	if len(key) > 0 && key[0] == '/' {
		key = key[1:]
	}
	if key == "" {
		response.Fail(c, 40002, "缺少文件 key")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	isAdmin := isAdmin(c)

	// 通过 storage_key 查找文件元数据
	fileMeta, err := h.svc.GetByStorageKey(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Body{
			Code:      1404,
			Message:   "文件不存在",
			RequestID: response.RequestIDFromContext(c),
		})
		return
	}

	reader, meta, err := h.svc.Download(c.Request.Context(), fileMeta.ID, userID, isAdmin)
	if err != nil {
		c.JSON(http.StatusForbidden, response.Body{
			Code:      1403,
			Message:   err.Error(),
			RequestID: response.RequestIDFromContext(c),
		})
		return
	}
	defer reader.Close()

	// 设置响应头
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, meta.OriginalName))
	c.Header("Content-Type", meta.MimeType)
	c.Header("Content-Length", fmt.Sprintf("%d", meta.SizeBytes))
	c.DataFromReader(http.StatusOK, meta.SizeBytes, meta.MimeType, reader, nil)
}

// Delete 软删除文件。DELETE /api/v1/files/:key
func (h *FileHandler) Delete(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.Fail(c, 40002, "缺少文件 key")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	isAdmin := isAdmin(c)

	// 通过 key 查找文件元数据
	fileMeta, err := h.svc.GetByStorageKey(c.Request.Context(), key)
	if err != nil {
		response.Fail(c, 1404, "文件不存在")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), fileMeta.ID, userID, isAdmin); err != nil {
		c.JSON(http.StatusForbidden, response.Body{
			Code:      1403,
			Message:   err.Error(),
			RequestID: response.RequestIDFromContext(c),
		})
		return
	}

	response.OK(c, gin.H{"key": key})
}

// GetMeta 获取文件元数据。GET /api/v1/files/:key/meta
func (h *FileHandler) GetMeta(c *gin.Context) {
	key := c.Param("key")
	// 兼容 /meta 后缀（虽然现在由 Download 中转调用，但保留以防直接调用）
	if strings.HasSuffix(key, "/meta") {
		key = strings.TrimSuffix(key, "/meta")
	}
	if key == "" {
		response.Fail(c, 40002, "缺少文件 key")
		return
	}

	fileMeta, err := h.svc.GetByStorageKey(c.Request.Context(), key)
	if err != nil {
		response.Fail(c, 1404, "文件不存在")
		return
	}

	response.OK(c, gin.H{
		"id":            fileMeta.ID,
		"key":           fileMeta.StorageKey,
		"original_name": fileMeta.OriginalName,
		"mime_type":     fileMeta.MimeType,
		"size":          fileMeta.SizeBytes,
		"hash":          "sha256:" + fileMeta.SHA256,
		"module":        fileMeta.Module,
		"biz_type":      fileMeta.BizType,
		"visibility":    fileMeta.Visibility,
		"uploader_id":   fileMeta.UploaderID,
		"created_at":    fileMeta.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
	})
}

// RegisterRoutes 注册文件模块路由。
func (h *FileHandler) RegisterRoutes(rg *gin.RouterGroup, adminOnly gin.HandlerFunc) {
	files := rg.Group("/files")
	{
		files.POST("/upload", h.Upload)
		// 使用通配符 *key 以支持含 "/" 的存储路径（如 2026/06/uuid.pdf）
		// /meta 路径在 Download handler 中识别，避免与 /*key 路由冲突
		files.GET("/*key", h.Download)
		files.DELETE("/*key", h.Delete)
	}
}

// isAdmin 判断当前用户是否为管理员。
func isAdmin(c *gin.Context) bool {
	roles, _ := c.Get("user_roles")
	roleList, _ := roles.([]string)
	for _, r := range roleList {
		if r == "R-SY-ADMIN" {
			return true
		}
	}
	return false
}

// isAllowedMIME 校验 MIME 类型是否在允许列表中。
func isAllowedMIME(mime string) bool {
	if mime == "" {
		return false
	}
	for _, allowed := range allowedMIMETypes {
		if strings.HasPrefix(mime, allowed) {
			return true
		}
	}
	return false
}
