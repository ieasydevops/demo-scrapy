<template>
  <el-container style="height: 100vh">
    <el-aside width="200px" style="background: #304156; color: white">
      <div style="padding: 20px; font-size: 18px; font-weight: bold; border-bottom: 1px solid #434a50">
        监控系统
      </div>
      <el-menu
        :default-active="activeMenu"
        @select="handleMenuSelect"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
        router
      >
        <el-sub-menu index="data-source">
          <template #title>
            <span>监控数据源管理</span>
          </template>
          <el-menu-item index="/web-pages">网页列表</el-menu-item>
          <el-menu-item index="/monitor-config">监控配置管理</el-menu-item>
          <el-menu-item index="/subscribe-config">订阅配置管理</el-menu-item>
        </el-sub-menu>
        <el-menu-item index="/announcements">采购信息动态</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="background: #409eff; color: white; display: flex; align-items: center; padding: 0 20px">
        <h1 style="margin: 0; font-size: 20px; font-weight: 500">政府采购网监控系统</h1>
      </el-header>
      <el-main style="padding: 20px; background: #f0f2f5">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export default {
  name: 'App',
  setup() {
    const route = useRoute()
    const router = useRouter()
    
    const activeMenu = computed(() => {
      const path = route.path
      if (path.startsWith('/web-pages')) return '/web-pages'
      if (path.startsWith('/monitor-config')) return '/monitor-config'
      if (path.startsWith('/subscribe-config')) return '/subscribe-config'
      if (path.startsWith('/announcements')) return '/announcements'
      return path
    })
    
    const handleMenuSelect = (path) => {
      router.push(path)
    }
    
    return {
      activeMenu,
      handleMenuSelect
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

.el-aside .el-menu {
  border-right: none;
}

.el-aside .el-menu-item {
  color: #bfcbd9;
}

.el-aside .el-menu-item:hover,
.el-aside .el-menu-item.is-active {
  background-color: #263445 !important;
  color: #409eff;
}

.el-aside .el-sub-menu__title:hover {
  background-color: #263445 !important;
}
</style>
