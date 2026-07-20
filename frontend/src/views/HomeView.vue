<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { generateProfile, listProfiles } from '../api'
import ProfileCard from '../components/ProfileCard.vue'

const router = useRouter()

const username = ref('')
const version = ref('outsider')
const loading = ref(false)
const loadingMsg = ref('')
const error = ref('')
const result = ref(null) // { id, profile, version }
const toast = ref('')
const history = ref(null) // null=加载中, []=空

const presets = [
  { key: 'insider', en: 'INSIDER BRIEF', cn: '圈内密报', desc: '写给同行看：行话黑话拉满，发技术社区、校友群，懂的人秒懂。' },
  { key: 'outsider', en: 'MATCHMAKING POST', cn: '相亲角快报', desc: '写给长辈和介绍人看：通俗接地气，突出生活魅力，发相亲群。' },
]

// ---- Hero 打字机 ----
const phrases = [
  '输入代号，破译 Ta 的互联网足迹',
  'GitHub · 博客 · 技术社区，足迹即线索',
  '圈内密报 或 相亲角快报，选一种语气',
]
const typedText = ref('')
let typeTimer = null

function startTypewriter() {
  let pi = 0, ci = 0, deleting = false
  typeTimer = setInterval(() => {
    const phrase = phrases[pi]
    if (!deleting) {
      typedText.value = phrase.slice(0, ++ci)
      if (ci === phrase.length) { deleting = true; ci = phrase.length + 24 } // 停顿
    } else if (--ci <= 0) {
      deleting = false; pi = (pi + 1) % phrases.length; ci = 0
    } else if (ci <= phrase.length) {
      typedText.value = phrase.slice(0, ci)
    }
  }, 70)
}

// ---- 加载文案轮播 ----
const loadingMsgs = [
  '正在检索公开足迹...',
  '正在分析语言风格...',
  '正在推断职业领域...',
  '正在撰写密报正文...',
]
let msgTimer = null

async function generate() {
  const u = username.value.trim()
  if (!u || loading.value) return
  loading.value = true
  error.value = ''
  result.value = null
  let mi = 0
  loadingMsg.value = loadingMsgs[0]
  msgTimer = setInterval(() => { loadingMsg.value = loadingMsgs[++mi % loadingMsgs.length] }, 4000)
  try {
    const data = await generateProfile(u, version.value)
    result.value = { ...data, version: version.value }
    loadHistory()
  } catch (e) {
    error.value = e.message
  } finally {
    clearInterval(msgTimer)
    loading.value = false
  }
}

async function loadHistory() {
  try {
    const data = await listProfiles()
    history.value = data.profiles || []
  } catch {
    history.value = []
  }
}

const PRESET_TAG = { insider: '圈内密报', outsider: '相亲角快报', both: '双版存档' }
function presetTag(v) { return PRESET_TAG[v] || v || '—' }

function fmtDate(s) {
  const d = new Date(s)
  return isNaN(d) ? '' : d.toLocaleDateString('zh-CN')
}

let toastTimer = null
function showToast(msg) {
  toast.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => (toast.value = ''), 2200)
}

onMounted(() => { startTypewriter(); loadHistory() })
onUnmounted(() => { clearInterval(typeTimer); clearInterval(msgTimer); clearTimeout(toastTimer) })
</script>

<template>
  <div>
    <div class="hero">
      <div class="kicker">// 互联网足迹破译计划</div>
      <h1>一个网名<br>读懂一个人</h1>
      <div class="typewriter">{{ typedText }}<span class="cursor"></span></div>
    </div>

    <div class="search-card">
      <div class="field-label">档案代号 / CODENAME</div>
      <div class="search-row">
        <input
          v-model="username"
          type="text"
          placeholder="输入网名或 GitHub 用户名..."
          autocomplete="off"
          autofocus
          @keydown.enter="generate"
        >
        <button class="btn-seal" :disabled="loading" @click="generate">
          {{ loading ? '破译中' : '开始破译 →' }}
        </button>
      </div>
      <div class="field-label" style="margin-top: 18px">选择生成预设 / PRESET（分享仅含所选版本）</div>
      <div class="preset-grid">
        <button
          v-for="p in presets"
          :key="p.key"
          class="preset-card"
          :class="{ active: version === p.key }"
          @click="version = p.key"
        >
          <span class="p-stamp">选用中</span>
          <div class="p-en">{{ p.en }}</div>
          <div class="p-cn">{{ p.cn }}</div>
          <div class="p-desc">{{ p.desc }}</div>
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading-box">
      <div class="ld-head">破译进行中 DECODING</div>
      <div class="ld-line">{{ loadingMsg }}</div>
      <div class="ld-bar"></div>
    </div>
    <div v-if="error" class="error-box">{{ error }}</div>

    <ProfileCard
      v-if="result"
      :profile="result.profile"
      :version="result.version"
      :share-id="result.id"
      @toast="showToast"
    />

    <section v-if="history === null || history.length" class="archive">
      <h2>档案柜 RECENT FILES</h2>
      <div v-if="history === null" class="archive-empty">调阅中...</div>
      <template v-else>
        <div
          v-for="item in history"
          :key="item.id"
          class="archive-row"
          @click="router.push(`/p/${item.id}`)"
        >
          <span class="a-no">No.{{ item.id.toUpperCase() }}</span>
          <span class="a-nick">{{ item.nickname || item.username }}</span>
          <span class="a-tag">{{ presetTag(item.version) }}</span>
          <span class="a-meta">{{ fmtDate(item.created_at) }} · 调阅 {{ item.views }}</span>
        </div>
      </template>
    </section>

    <div class="toast" :class="{ show: !!toast }">{{ toast }}</div>
  </div>
</template>
