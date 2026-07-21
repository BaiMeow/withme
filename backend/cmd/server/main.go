package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"withme/internal/api"
	"withme/internal/config"
	"withme/internal/generator"
	"withme/internal/moderation"
	"withme/internal/store"
	"withme/internal/web"
)

func main() {
	cfg := loadConfig()

	st, err := store.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer st.Close()
	log.Printf("[database] driver=%s ready\n", cfg.Database.Driver)

	gen, err := generator.New(context.Background(), cfg.Gemini.APIKey, cfg.Gemini.Model)
	if err != nil {
		log.Fatalf("failed to init gemini: %v", err)
	}
	log.Printf("[gemini] model=%s ready\n", cfg.Gemini.Model)

	mod, err := moderation.New(cfg.Tencent.SecretID, cfg.Tencent.SecretKey, cfg.Tencent.Region, cfg.Tencent.BizType)
	if err != nil {
		log.Fatalf("failed to init tencent tms: %v", err)
	}
	if mod.Enabled() {
		log.Printf("[moderation] tencent tms ready (region=%s biz_type=%s)\n", cfg.Tencent.Region, cfg.Tencent.BizType)
	} else {
		log.Println("[moderation] tencent.secret_id/secret_key 未配置，内容审核已关闭")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	handler := api.NewHandler(gen, st, mod)
	r.POST("/api/generate", handler.GenerateProfile)
	r.GET("/api/profiles", handler.ListProfiles)
	r.GET("/api/profiles/:id", handler.GetProfile)

	// 前端产物经 go:embed 内嵌：静态资源 + SPA 回退（/p/:id 等前端路由）
	dist, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		log.Fatalf("failed to load embedded frontend: %v", err)
	}
	index, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		log.Println("[frontend] 未找到内嵌的前端产物，请先执行 cd frontend && pnpm build（仅 API 可用）")
	} else {
		assets, _ := fs.Sub(dist, "assets")
		r.StaticFS("/assets", http.FS(assets))
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", index)
		})
		log.Println("[frontend] embedded dist ready")
	}

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("[server] listening on %s\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() *config.Config {
	cfg, err := config.Load("./config/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if cfg.Gemini.APIKey == "" {
		log.Fatal("gemini.api_key is required in config/config.yaml")
	}
	return cfg
}
