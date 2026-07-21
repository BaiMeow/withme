package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"withme/internal/generator"
	"withme/internal/model"
	"withme/internal/moderation"
	"withme/internal/store"
)

type Handler struct {
	generator *generator.Generator
	store     *store.Store
	moderator *moderation.Moderator
}

func NewHandler(gen *generator.Generator, st *store.Store, mod *moderation.Moderator) *Handler {
	return &Handler{generator: gen, store: st, moderator: mod}
}

// GenerateProfile POST /api/generate 生成资料并入库，返回分享 ID。
func (h *Handler) GenerateProfile(c *gin.Context) {
	var req model.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.GenerateResponse{Error: "请提供 username 字段"})
		return
	}
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, model.GenerateResponse{Error: "username 不能为空"})
		return
	}
	// 赛博版 / 相亲角版是两个独立预设，一次只生成一个版本
	if req.Version == "" {
		req.Version = "normal"
	}
	if req.Version != "cyber" && req.Version != "normal" {
		c.JSON(http.StatusBadRequest, model.GenerateResponse{Error: "version 只能是 cyber（赛博版）或 normal（相亲角版）"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	profile, err := h.generator.Generate(ctx, req.Username, req.Version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.GenerateResponse{Error: "生成失败: " + err.Error()})
		return
	}

	// 生成内容对外可见，入库前过一遍内容安全审核；
	// 审核服务异常时放行（fail-open），避免腾讯云故障拖垮生成链路
	if h.moderator.Enabled() {
		reason, err := h.moderator.Check(ctx, profileText(req.Username, profile))
		if err != nil {
			slog.Warn("moderation check failed, fail-open", "error", err)
		} else if reason != "" {
			slog.Info("moderation blocked", "username", req.Username, "reason", reason)
			c.JSON(http.StatusOK, model.GenerateResponse{Error: "生成内容未通过安全审核，请换个用户名重试"})
			return
		}
	}

	id, err := h.store.Save(ctx, req.Username, req.Version, profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.GenerateResponse{Error: "保存失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.GenerateResponse{ID: id, Profile: profile})
}

// profileText 拼接对外展示的全部文本，一次调用完成审核
func profileText(username string, p *model.DatingProfile) string {
	var sb strings.Builder
	sb.WriteString(username + "\n" + p.Nickname + "\n")
	sb.WriteString(p.BasicInfo.Gender + " " + p.BasicInfo.AgeRange + " " + p.BasicInfo.Location + " " + p.BasicInfo.Occupation + "\n")
	sb.WriteString(p.Content)
	return sb.String()
}

// GetProfile GET /api/profiles/:id 按分享 ID 取资料（浏览数 +1）。
func (h *Handler) GetProfile(c *gin.Context) {
	sp, err := h.store.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "资料不存在或已被删除"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, sp)
}

// ListProfiles GET /api/profiles 最近生成的资料摘要。
func (h *Handler) ListProfiles(c *gin.Context) {
	list, err := h.store.ListRecent(c.Request.Context(), 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profiles": list})
}
