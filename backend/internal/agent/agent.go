// Package agent 在 genai Chat（自动管理历史）之上提供带工具调用的 Agent 循环：
// 模型发起 function call 时执行本地注册的工具，把结果回喂给模型，直到输出文本或超过轮次上限。
// Google Search 等服务端工具无需本地执行，由模型侧直接调用。
package agent

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/genai"
)

// Tool 本地工具。Declaration 告诉模型工具怎么调，Call 真正执行。
// Call 的返回值会被 JSON 序列化后作为 function response 回喂给模型。
type Tool interface {
	Declaration() *genai.FunctionDeclaration
	Call(ctx context.Context, args map[string]any) (any, error)
}

type Agent struct {
	client   *genai.Client
	model    string
	config   *genai.GenerateContentConfig
	tools    map[string]Tool
	maxTurns int
}

type Option func(*Agent)

// WithTools 注册本地工具，同名后注册覆盖先注册
func WithTools(tools ...Tool) Option {
	return func(a *Agent) {
		for _, t := range tools {
			a.tools[t.Declaration().Name] = t
		}
	}
}

// WithGoogleSearch 启用 Google Search grounding（服务端工具，无需本地实现）
func WithGoogleSearch() Option {
	return func(a *Agent) {
		a.config.Tools = append(a.config.Tools, &genai.Tool{GoogleSearch: &genai.GoogleSearch{}})
	}
}

// WithConfig 覆盖生成配置（温度等）；工具声明在 Run 时组装，此处无需设置 Tools
func WithConfig(cfg *genai.GenerateContentConfig) Option {
	return func(a *Agent) {
		if cfg != nil {
			a.config = cfg
		}
	}
}

// WithMaxTurns 限制工具调用轮次，防止模型陷入调用循环，默认 8
func WithMaxTurns(n int) Option {
	return func(a *Agent) {
		if n > 0 {
			a.maxTurns = n
		}
	}
}

func New(client *genai.Client, model string, opts ...Option) *Agent {
	a := &Agent{
		client:   client,
		model:    model,
		config:   &genai.GenerateContentConfig{},
		tools:    map[string]Tool{},
		maxTurns: 8,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Run 执行一轮完整任务：从 prompt 开始，直到模型给出不含工具调用的文本回复
func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
	cfg := *a.config
	if len(a.tools) > 0 {
		decls := make([]*genai.FunctionDeclaration, 0, len(a.tools))
		for _, t := range a.tools {
			decls = append(decls, t.Declaration())
		}
		cfg.Tools = append(cfg.Tools, &genai.Tool{FunctionDeclarations: decls})
	}
	// debug 级别才请求思考摘要，避免平时为 thinking token 额外付费
	debug := slog.Default().Enabled(ctx, slog.LevelDebug)
	if debug {
		cfg.ThinkingConfig = &genai.ThinkingConfig{IncludeThoughts: true}
	}

	chat, err := a.client.Chats.Create(ctx, a.model, &cfg, nil)
	if err != nil {
		return "", fmt.Errorf("create chat: %w", err)
	}

	parts := []*genai.Part{{Text: prompt}}
	for turn := 0; turn < a.maxTurns; turn++ {
		resp, err := chat.Send(ctx, parts...)
		if err != nil {
			return "", fmt.Errorf("gemini call failed: %w", err)
		}
		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			return "", fmt.Errorf("gemini returned empty response")
		}

		var calls []*genai.FunctionCall
		var text string
		for _, part := range resp.Candidates[0].Content.Parts {
			switch {
			case part.Thought:
				if debug {
					slog.DebugContext(ctx, "gemini thought", "thought", part.Text)
				}
			case part.FunctionCall != nil:
				calls = append(calls, part.FunctionCall)
			case text == "" && part.Text != "":
				text = part.Text
			}
		}

		if len(calls) == 0 {
			if text == "" {
				return "", fmt.Errorf("gemini returned empty response")
			}
			return text, nil
		}

		// 必须换新切片：chat 历史持有 parts 的引用，原地复用会篡改首条 user 消息
		responses := make([]*genai.Part, 0, len(calls))
		for _, call := range calls {
			responses = append(responses, a.execute(ctx, call))
		}
		parts = responses
	}
	return "", fmt.Errorf("agent exceeded max turns (%d), last task unfinished", a.maxTurns)
}

// execute 执行一次工具调用并包装成 function response part；
// 工具不存在或执行失败不中断，把错误回喂给模型自行补救
func (a *Agent) execute(ctx context.Context, call *genai.FunctionCall) *genai.Part {
	tool, ok := a.tools[call.Name]
	if !ok {
		slog.WarnContext(ctx, "gemini called unknown tool", "tool", call.Name)
		return genai.NewPartFromFunctionResponse(call.Name, map[string]any{"error": "unknown tool: " + call.Name})
	}

	slog.DebugContext(ctx, "tool call", "tool", call.Name, "args", call.Args)
	result, err := tool.Call(ctx, call.Args)
	if err != nil {
		slog.WarnContext(ctx, "tool call failed", "tool", call.Name, "error", err)
		return genai.NewPartFromFunctionResponse(call.Name, map[string]any{"error": err.Error()})
	}
	slog.DebugContext(ctx, "tool result", "tool", call.Name, "result", result)
	return genai.NewPartFromFunctionResponse(call.Name, map[string]any{"output": result})
}
