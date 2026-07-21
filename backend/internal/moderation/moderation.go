// Package moderation 封装腾讯云内容安全（TMS）文本审核。
// 未配置密钥时 Enabled() 为 false，调用方应直接放行。
package moderation

import (
	"context"
	"encoding/base64"
	"fmt"
	"unicode/utf8"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tms/v20201229"
)

// TMS Content 上限约 1 万字节（base64 前），留余量截断
const maxContentBytes = 9000

type Moderator struct {
	client  *tms.Client
	bizType string
}

// New 创建审核器；secretID/secretKey 为空时返回未启用状态（Enabled() == false）
func New(secretID, secretKey, region, bizType string) (*Moderator, error) {
	if secretID == "" || secretKey == "" {
		return &Moderator{}, nil
	}
	credential := common.NewCredential(secretID, secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "tms.tencentcloudapi.com"
	client, err := tms.NewClient(credential, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("create tms client: %w", err)
	}
	return &Moderator{client: client, bizType: bizType}, nil
}

// Enabled 是否已配置密钥、真正执行审核
func (m *Moderator) Enabled() bool {
	return m != nil && m.client != nil
}

// Check 审核文本。返回非空 reason 表示命中违规（建议拦截）；
// err 非空表示审核服务调用失败，由调用方决定放行策略。
func (m *Moderator) Check(ctx context.Context, text string) (reason string, err error) {
	text = truncate(text, maxContentBytes)

	request := tms.NewTextModerationRequest()
	request.SetContext(ctx)
	request.Content = common.StringPtr(base64.StdEncoding.EncodeToString([]byte(text)))
	request.BizType = common.StringPtr(m.bizType)

	response, err := m.client.TextModeration(request)
	if err != nil {
		return "", fmt.Errorf("tms TextModeration: %w", err)
	}

	r := response.Response
	if r.Suggestion == nil || *r.Suggestion == "Pass" {
		return "", nil
	}
	// Block / Review 均视为不合规，附上级标签便于排查
	label, subLabel := "", ""
	if r.Label != nil {
		label = *r.Label
	}
	if r.SubLabel != nil {
		subLabel = *r.SubLabel
	}
	return fmt.Sprintf("suggestion=%s label=%s sub_label=%s", *r.Suggestion, label, subLabel), nil
}

// truncate 按 UTF-8 字符边界截断到 max 字节以内
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	for max > 0 && !utf8.ValidString(s[:max]) {
		max--
	}
	return s[:max]
}
