package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"withme/internal/model"

	"google.golang.org/genai"
)

type Generator struct {
	client *genai.Client
	model  string
}

func New(ctx context.Context, apiKey, model string) (*Generator, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &Generator{client: client, model: model}, nil
}

func (g *Generator) Generate(ctx context.Context, username, version string) (*model.DatingProfile, error) {
	cfg := &genai.GenerateContentConfig{
		Temperature: genai.Ptr[float32](0.8),
		Tools: []*genai.Tool{
			{GoogleSearch: &genai.GoogleSearch{}},
		},
	}

	// debug 级别才请求思考摘要，避免平时为 thinking token 额外付费
	debug := slog.Default().Enabled(ctx, slog.LevelDebug)
	if debug {
		cfg.ThinkingConfig = &genai.ThinkingConfig{IncludeThoughts: true}
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(pickPrompt(version, username)), cfg)
	if err != nil {
		return nil, fmt.Errorf("gemini call failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned empty response")
	}

	var content string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Thought {
			if debug {
				slog.DebugContext(ctx, "gemini thought", "username", username, "version", version, "thought", part.Text)
			}
			continue
		}
		if content == "" && part.Text != "" {
			content = part.Text
		}
	}
	if content == "" {
		return nil, fmt.Errorf("gemini returned empty response")
	}
	content = extractJSON(content)

	var profile model.DatingProfile
	if err := json.Unmarshal([]byte(content), &profile); err != nil {
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
