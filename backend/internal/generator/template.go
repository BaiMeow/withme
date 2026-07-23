package generator

import "fmt"

// pickPrompt 通用信息 + 版本特殊要求拼装；version 为 cyber / normal
// 两个预设共享同一段检索与输出要求，每次只生成一个版本
func pickPrompt(version, username string) string {
	specific := promptNormal
	if version == "cyber" {
		specific = promptCyber
	}
	return fmt.Sprintf(promptCommon, username, username, username, specific)
}

const promptCommon = `目标：为网名为 "%s" 的人生成一份相亲资料。你必须调用所有可用的函数，从多个平台、多个角度全面采集信息。禁止仅依赖单一来源或跳过任何检索步骤。

第一步·Google 搜索定位（必须执行）：
用 "%s" 进行 Google 搜索，在微博、Twitter、GitHub、知乎等公开社交平台中定位到最匹配的人，收集 TA 的各种别名——曾用名、其他平台账号名、Steam 昵称等。若搜索结果明显指向多个不同的人，以信息最丰富、最活跃的那个为准。

第二步·X(Twitter) 信息采集（若有可用工具则必须执行）：
	先用 x_search_user 工具搜索 Google 中找到的 Twitter/X 用户名，获取用户ID。成功后立即调用 x_userpost 工具获取该用户所有近期推文，分析推文内容以了解其兴趣爱好、日常关注话题、社交圈层和语言风格。两个工具应联合使用，不可只调用其中一个。

	第三步·Steam 信息采集（必须执行）：
调用 steam_search_users 工具，用上一步找到的 Steam 昵称（如无则直接用 "%s"）搜索 Steam 社区用户。对搜索结果中匹配度最高的几个用户，逐一调用 steam_get_profile 工具抓取其详细信息（昵称、等级、在线状态、地区、简介、最近游玩的游戏与时长、成就进度、展柜游戏、愿望单数、加入的组等），交叉比对头像、地区、简介等确认目标。

	第四步·交叉验证：
	综合 Google 搜索、X(Twitter) 推文和 Steam 个人主页详情，互相印证，去伪存真。信息冲突时以各平台官方数据为准，社交平台内容作为补充。

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
