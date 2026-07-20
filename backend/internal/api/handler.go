package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"withme/internal/generator"
	"withme/internal/model"
	"withme/internal/store"
)

type Handler struct {
	generator *generator.Generator
	store     *store.Store
}

func NewHandler(gen *generator.Generator, st *store.Store) *Handler {
	return &Handler{generator: gen, store: st}
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
	// 圈内人版 / 圈外人版是两个独立预设，一次只生成一个版本
	if req.Version == "" {
		req.Version = "outsider"
	}
	if req.Version != "insider" && req.Version != "outsider" {
		c.JSON(http.StatusBadRequest, model.GenerateResponse{Error: "version 只能是 insider（圈内人版）或 outsider（圈外人版）"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	profile, err := h.generator.Generate(ctx, req.Username, req.Version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.GenerateResponse{Error: "生成失败: " + err.Error()})
		return
	}

	id, err := h.store.Save(ctx, req.Username, req.Version, profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.GenerateResponse{Error: "保存失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.GenerateResponse{ID: id, Profile: profile})
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
