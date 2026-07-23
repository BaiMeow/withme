package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"withme/internal/crawler/steam"

	"withme/internal/agent"
	"withme/internal/crawler/x"
	"withme/internal/model"

	"google.golang.org/genai"
)

type Generator struct {
	agent *agent.Agent
}

// New 创建资料生成器；xBearerToken 非空时启用 x_search_user / x_userpost 工具
func New(ctx context.Context, apiKey, model, xBearerToken string) (*Generator, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	steamSearch, steamProfile := steam.NewTools()

	tools := []agent.Tool{steamSearch, steamProfile}
	if xBearerToken != "" {
		xSearch, xPost := x.NewTools(xBearerToken)
		tools = append(tools, xSearch, xPost)
	}

	a := agent.New(client, model,
		agent.WithGoogleSearch(),
		agent.WithTools(tools...),
	)
	return &Generator{agent: a}, nil
}

func (g *Generator) Generate(ctx context.Context, username, version string) (*model.DatingProfile, error) {
	content, err := g.agent.Run(ctx, pickPrompt(version, username))
	if err != nil {
		return nil, err
	}

	var profile model.DatingProfile
	if err := json.Unmarshal([]byte(extractJSON(content)), &profile); err != nil {
		return nil, fmt.Errorf("failed to parse gemini output: %w\nraw: %s", err, content)
	}

	return &profile, nil
}

func extractJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, "{"); idx != -1 {
		raw = raw[idx:]
	}
	if idx := strings.LastIndex(raw, "}"); idx != -1 {
		raw = raw[:idx+1]
	}
	return raw
}
