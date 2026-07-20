<template>
  <div class="app-shell">
    <header class="app-header">
      <router-link to="/" class="logo">
        <span class="logo-cn">相亲档案室</span>
        <span class="logo-en">WITHME·ARCHIVE</span>
      </router-link>
      <router-link v-if="$route.name !== 'home'" to="/" class="btn-link">建立新档案 →</router-link>
    </header>
    <main class="app-main">
      <router-view />
    </main>
    <footer class="app-footer">WITHME ARCHIVE · 仅整理公开信息 · AI 生成内容仅供参考</footer>
  </div>
</template>
