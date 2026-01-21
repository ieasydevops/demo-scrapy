<template>
  <div>
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>订阅配置管理</span>
          <el-button type="primary" @click="showDialog = true">添加订阅</el-button>
        </div>
      </template>

      <el-table :data="configs" border v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="email" label="邮箱地址" />
        <el-table-column prop="push_time" label="推送时间" />
        <el-table-column prop="created_at" label="创建时间" />
        <el-table-column label="操作" width="180">
          <template #default="scope">
            <el-button size="small" @click="editConfig(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteConfig(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑订阅' : '添加订阅'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="邮箱地址" required>
          <el-input v-model="form.email" placeholder="请输入邮箱地址" />
        </el-form-item>
        <el-form-item label="推送时间" required>
          <el-time-picker
            v-model="pushTime"
            format="HH:mm"
            value-format="HH:mm"
            placeholder="选择推送时间"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="saveConfig">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSubscribeConfig, createSubscribeConfig, updateSubscribeConfig, deleteSubscribeConfig } from '../api'

export default {
  name: 'SubscribeConfig',
  setup() {
    const configs = ref([])
    const showDialog = ref(false)
    const editingId = ref(null)
    const loading = ref(false)
    const pushTime = ref('')
    const form = ref({
      email: '',
      push_time: ''
    })

    const loadConfigs = async () => {
      loading.value = true
      try {
        const res = await getSubscribeConfig()
        configs.value = res.data
      } catch (error) {
        console.error('加载失败:', error)
        ElMessage.error('加载失败')
      } finally {
        loading.value = false
      }
    }

    const editConfig = (row) => {
      editingId.value = row.id
      form.value = {
        email: row.email,
        push_time: row.push_time
      }
      pushTime.value = row.push_time
      showDialog.value = true
    }

    const saveConfig = async () => {
      if (!form.value.email || !pushTime.value) {
        ElMessage.warning('请填写完整信息')
        return
      }

      form.value.push_time = pushTime.value

      try {
        if (editingId.value) {
          await updateSubscribeConfig(editingId.value, form.value)
          ElMessage.success('更新成功')
        } else {
          await createSubscribeConfig(form.value)
          ElMessage.success('添加成功')
        }
        showDialog.value = false
        editingId.value = null
        form.value = { email: '', push_time: '' }
        pushTime.value = ''
        loadConfigs()
      } catch (error) {
        console.error('操作失败:', error)
        const message = error.response?.data?.error || error.message || '操作失败'
        ElMessage.error(message)
      }
    }

    const deleteConfig = async (id) => {
      try {
        await deleteSubscribeConfig(id)
        ElMessage.success('删除成功')
        loadConfigs()
      } catch (error) {
        console.error('删除失败:', error)
        ElMessage.error('删除失败')
      }
    }

    onMounted(() => {
      loadConfigs()
    })

    return {
      configs,
      showDialog,
      editingId,
      loading,
      pushTime,
      form,
      editConfig,
      saveConfig,
      deleteConfig
    }
  }
}
</script>
