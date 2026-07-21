<script setup>
import { computed } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const props = defineProps({
  profile: { type: Object, required: true }, // DatingProfile
  version: { type: String, default: 'normal' }, // 生成预设：cyber / normal
  shareId: { type: String, default: '' },
  views: { type: Number, default: null },
  createdAt: { type: String, default: '' },
})

const emit = defineEmits(['toast'])

const PRESETS = {
  cyber: { cn: '圈内密报', en: 'INSIDER BRIEF' },
  normal: { cn: '相亲角快报', en: 'MATCHMAKING POST' },
}

// version 非法时回退 normal
const effectiveVersion = computed(() => (PRESETS[props.version] ? props.version : 'normal'))

const preset = computed(() => PRESETS[effectiveVersion.value])

const content = computed(() => props.profile.content || '')

const contentHtml = computed(() =>
  content.value ? DOMPurify.sanitize(marked.parse(content.value)) : '',
)

const fileNo = computed(() => (props.shareId ? props.shareId.toUpperCase() : 'PENDING'))

const createdAtText = computed(() => {
  if (!props.createdAt) return ''
  const d = new Date(props.createdAt)
  return isNaN(d) ? '' : d.toLocaleString('zh-CN', { hour12: false })
})

const infoRows = computed(() => {
  const b = props.profile.basic_info || {}
  return [
    ['代号 CODENAME', props.profile.nickname || '未知'],
    ['性别 GENDER', b.gender || '未知'],
    ['年龄段 AGE', b.age_range || '未知'],
    ['所在地 LOCATION', b.location || '未知'],
    ['领域 OCCUPATION', b.occupation || '未知'],
  ]
})

async function copyShareLink() {
  const url = `${location.origin}/p/${props.shareId}`
  try {
    await navigator.clipboard.writeText(url)
    emit('toast', '密件链接已复制')
  } catch {
    // 剪贴板不可用（如非 https），降级为手动复制
    window.prompt('复制密件链接：', url)
  }
}
</script>

<template>
  <div class="dossier">
    <div class="dossier-head">
      <span class="file-no">档案编号 NO.{{ fileNo }}</span>
      <span>密级：公开资料整理</span>
    </div>

    <div class="dossier-card">
      <table class="info-table">
        <tbody>
          <tr v-for="([k, v], i) in infoRows" :key="k">
            <td class="k">{{ k }}</td>
            <td class="v">
              <span class="reveal" :style="{ '--i': i }">{{ v }}</span>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="seal-stamp">
        <span>已破译</span>
        <span class="s-en">DECODED</span>
      </div>
    </div>

    <div class="dossier-card">
      <span class="preset-tag">{{ preset.cn }} · {{ preset.en }}</span>
      <div class="md-body" v-html="contentHtml"></div>
    </div>

    <div v-if="profile.sources?.length" class="sources-row">
      线索来源：
      <a v-for="(s, i) in profile.sources" :key="i" :href="s" target="_blank" rel="noopener">[{{ i + 1 }}]</a>
    </div>

    <div class="dossier-actions">
      <button v-if="shareId" class="btn-seal" @click="copyShareLink">🔒 封存并复制密件链接</button>
      <span v-if="views !== null" class="meta-mono">调阅 {{ views }} 次</span>
      <span v-if="createdAtText" class="meta-mono">归档于 {{ createdAtText }}</span>
    </div>
  </div>
</template>

<style scoped>
/* 每行脱敏条依次揭开 */
.reveal::after { animation-delay: calc(var(--i, 0) * 0.12s); }
</style>
