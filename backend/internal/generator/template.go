package generator

import "fmt"

// pickPrompt 根据版本选择 prompt，username 作为搜索关键词
// 圈内人版 / 圈外人版是两个独立预设，每次只生成一个版本
func pickPrompt(version, username string) string {
	if version == "insider" {
		return fmt.Sprintf(promptInsider, username)
	}
	return fmt.Sprintf(promptOutsider, username)
}

const promptInsider = `请使用 Google 搜索网名为 "%s" 的人在互联网上的公开信息（GitHub、博客、技术社区、社交媒体等），全面了解其职业领域、技术方向、兴趣圈子和语言风格，然后生成一份相亲资料。

只生成极客版——适合发在同行圈子、技术社区、校友群。用该领域的行话和圈子文化来写，让同行看了会心一笑。

严格按以下 JSON 输出（不要带其他文字）：

{
  "nickname": "网名",
  "basic_info": {
    "gender": "推测性别",
    "age_range": "推测年龄段",
    "location": "所在地",
    "occupation": "职业领域"
  },
  "insider": "极客版完整内容（markdown格式）",
  "sources": ["信息源URL"]
}`

const promptOutsider = `请使用 Google 搜索网名为 "%s" 的人在互联网上的公开信息（GitHub、博客、技术社区、社交媒体等），全面了解其职业领域、兴趣圈子和生活方式，然后生成一份相亲资料。

只生成接地气版——适合发相亲角、相亲群、给长辈看。完全用通俗语言，去掉所有行业术语，突出生活化魅力。

严格按以下 JSON 输出（不要带任何其他文字）：

{
  "nickname": "网名",
  "basic_info": {
    "gender": "推测性别",
    "age_range": "推测年龄段",
    "location": "所在地",
    "occupation": "职业领域"
  },
  "outsider": "接地气版完整内容（markdown格式）",
  "sources": ["信息源URL"]
}`
