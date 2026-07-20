package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"withme/internal/api"
	"withme/internal/config"
	"withme/internal/generator"
	"withme/internal/store"
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

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	handler := api.NewHandler(gen, st)
	r.POST("/api/generate", handler.GenerateProfile)
	r.GET("/api/profiles", handler.ListProfiles)
	r.GET("/api/profiles/:id", handler.GetProfile)

	// 托管前端构建产物：静态资源 + SPA 回退（/p/:id 等前端路由）
	dist := cfg.Frontend.Dist
	if _, err := os.Stat(filepath.Join(dist, "index.html")); err != nil {
		log.Printf("[frontend] %s 不存在，请先执行 npm run build（仅 API 可用）\n", dist)
	} else {
		r.Static("/assets", filepath.Join(dist, "assets"))
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.File(filepath.Join(dist, "index.html"))
		})
		log.Printf("[frontend] serving %s\n", dist)
	}

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("[server] listening on %s\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() *config.Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "./config/config.yaml"
	}
	cfg, err := config.Load(path)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if cfg.Gemini.APIKey == "" {
		log.Fatal("gemini.api_key is required in config/config.yaml")
	}
	return cfg
}
