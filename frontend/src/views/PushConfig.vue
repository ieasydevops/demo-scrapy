<template>
  <div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="推送配置" name="config">
        <el-form :model="config" label-width="120px" style="max-width: 500px">
          <el-form-item label="邮箱地址">
            <el-input v-model="config.email" />
          </el-form-item>
          <el-form-item label="推送时间">
            <el-input-number v-model="hour" :min="0" :max="23" />
            <span style="margin-left: 10px">点（24小时制）</span>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveConfig">保存配置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getPushConfig, updatePushConfig } from '../api'

export default {
  name: 'PushConfig',
  setup() {
    const config = ref({ email: '', push_time: '' })
    const hour = ref(17)
    const activeTab = ref('config')

    const loadConfig = async () => {
      try {
        const res = await getPushConfig()
        config.value = res.data
        if (res.data.push_time) {
          hour.value = parseInt(res.data.push_time)
        }
      } catch (error) {
        console.error('加载失败:', error)
        const message = error.response?.data?.error || error.message || '加载失败'
        ElMessage.error(message)
      }
    }

    const saveConfig = async () => {
      if (!config.value.email) {
        ElMessage.warning('请输入邮箱地址')
        return
      }
      try {
        await updatePushConfig({ email: config.value.email, push_time: hour.value.toString() })
        ElMessage.success('保存成功')
      } catch (error) {
        console.error('保存失败:', error)
        const message = error.response?.data?.error || error.message || '保存失败'
        ElMessage.error(message)
      }
    }

    onMounted(() => {
      loadConfig()
    })

    return {
      config,
      hour,
      activeTab,
      saveConfig
    }
  }
}
</script>
