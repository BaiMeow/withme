package generator

import "fmt"

// pickPrompt 通用信息 + 版本特殊要求拼装；version 为 cyber / normal
// 两个预设共享同一段检索与输出要求，每次只生成一个版本
func pickPrompt(version, username string) string {
	specific := promptNormal
	if version == "cyber" {
		specific = promptCyber
	}
	return fmt.Sprintf(promptCommon, username, username, specific)
}

const promptCommon = `目标：为网名为 "%s" 的人生成一份相亲资料。

第一步·定位到人：
先用 "%s" 做初步检索（Google 搜索、Steam 用户搜索等可用工具），交叉比对各平台信息（头像、地区、简介、互相关联的账号等）定位到具体的人，并尽可能收集 TA 的各种别名——曾用名、其他平台账号名、昵称变体。若搜索结果明显指向多个不同的人，以信息最丰富、最活跃的那个为准。

第二步·详细检索：
用上一步确认到的所有别名，分别在各自活跃的平台上做详细检索，全面了解 TA 的性格特点、兴趣爱好、生活方式与网络形象。

%s

严格按以下 JSON 输出（不要带任何其他文字，不要用代码块包裹）：

{
  "nickname": "网名",
  "basic_info": {
    "gender": "性别",
    "age_range": "年龄段",
    "location": "所在地",
    "occupation": "职业领域"
  },
  "content": "正文（markdown格式）",
  "sources": ["信息源URL"]
}`

const promptCyber = `写作要求（赛博版）：
写给网友和同好看——适合发在 TA 所在的社区、群聊。用网络流行语和 TA 实际所在圈子的语言来写：TA 混游戏圈就用游戏梗，混动漫圈就用动漫梗，按检索到的真实圈子灵活调整，不要预设 TA 是程序员，互联网上不只有程序员。像朋友安利一样，还原 TA 在网上真实的样子，让同好看了会心一笑。`

const promptNormal = `写作要求（相亲角版）：
写给长辈和介绍人看——适合发相亲角、相亲群。完全用通俗语言，要接地气。末尾必须附一段「红娘点评」，以红娘的口吻总结此人的亮点，并给出一句真诚的推荐语。`
