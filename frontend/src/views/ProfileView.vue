<script setup>
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getProfile } from '../api'
import ProfileCard from '../components/ProfileCard.vue'

const route = useRoute()

const loading = ref(true)
const error = ref('')
const data = ref(null) // StoredProfile
const toast = ref('')

let toastTimer = null
function showToast(msg) {
  toast.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => (toast.value = ''), 2200)
}

async function load(id) {
  loading.value = true
  error.value = ''
  data.value = null
  try {
    data.value = await getProfile(id)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(() => load(route.params.id))
watch(() => route.params.id, (id) => id && load(id))
</script>

<template>
  <div>
    <div class="share-head">
      <div class="wax">密</div>
      <h1>你收到一份密件</h1>
      <div class="kicker">SEALED FILE · 经互联网公开足迹整理</div>
    </div>

    <div v-if="loading" class="loading-box">
      <div class="ld-head">拆封中 UNSEALING</div>
      <div class="ld-line">正在拆封档案...</div>
      <div class="ld-bar"></div>
    </div>
    <div v-else-if="error" class="error-box">{{ error }}</div>

    <template v-else-if="data">
      <ProfileCard
        :profile="data.profile"
        :version="data.version"
        :share-id="data.id"
        :views="data.views"
        :created-at="data.created_at"
        @toast="showToast"
      />
      <div class="share-actions">
        <router-link to="/" class="btn-link">输入网名，破译你自己的档案 →</router-link>
      </div>
    </template>

    <div class="toast" :class="{ show: !!toast }">{{ toast }}</div>
  </div>
</template>
