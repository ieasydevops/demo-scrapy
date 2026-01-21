<template>
  <div>
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>监控配置管理</span>
          <el-button type="primary" @click="handleAdd">添加配置</el-button>
        </div>
      </template>

      <el-table :data="configs" border v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="web_page_name" label="网页名称" />
        <el-table-column prop="crawl_time" label="采集时间" />
        <el-table-column prop="crawl_freq" label="采集频率" />
        <el-table-column label="关键字">
          <template #default="scope">
            <el-tag v-for="(kw, idx) in scope.row.keywords" :key="idx" style="margin-right: 5px">
              {{ kw }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" />
        <el-table-column label="操作" width="180">
          <template #default="scope">
            <el-button size="small" @click="editConfig(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteConfig(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑配置' : '添加配置'" width="600px">
      <el-form :model="form" label-width="120px">
        <el-form-item label="网页" required>
          <el-select v-model="form.web_page_id" placeholder="选择网页" style="width: 100%">
            <el-option
              v-for="page in webPages"
              :key="page.id"
              :label="page.name"
              :value="page.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="采集时间" required>
          <el-time-picker
            v-model="crawlTime"
            format="HH:mm"
            value-format="HH:mm"
            placeholder="选择时间"
            style="width: 100%"
            :clearable="true"
          />
        </el-form-item>
        <el-form-item label="采集频率" required>
          <el-select v-model="form.crawl_freq" placeholder="选择频率" style="width: 100%">
            <el-option label="每天" value="daily" />
            <el-option label="每小时" value="hourly" />
            <el-option label="每30分钟" value="30min" />
            <el-option label="每15分钟" value="15min" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键字" required>
          <el-select
            v-model="form.keywords"
            multiple
            filterable
            allow-create
            placeholder="选择或输入关键字"
            style="width: 100%"
          >
            <el-option
              v-for="kw in availableKeywords"
              :key="kw.id"
              :label="kw.keyword"
              :value="kw.keyword"
            />
          </el-select>
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
import { getMonitorConfig, createMonitorConfig, updateMonitorConfig, deleteMonitorConfig } from '../api'
import { getWebPages } from '../api'
import { getKeywords } from '../api'

export default {
  name: 'MonitorConfig',
  setup() {
    const configs = ref([])
    const webPages = ref([])
    const availableKeywords = ref([])
    const showDialog = ref(false)
    const editingId = ref(null)
    const loading = ref(false)
    const crawlTime = ref('')
    const form = ref({
      web_page_id: null,
      crawl_time: '',
      crawl_freq: '',
      keywords: []
    })

    const loadConfigs = async () => {
      loading.value = true
      try {
        const res = await getMonitorConfig()
        configs.value = res.data
      } catch (error) {
        console.error('加载失败:', error)
        ElMessage.error('加载失败')
      } finally {
        loading.value = false
      }
    }

    const loadWebPages = async () => {
      try {
        const res = await getWebPages()
        webPages.value = res.data
      } catch (error) {
        console.error('加载网页列表失败:', error)
      }
    }

    const loadKeywords = async () => {
      try {
        const res = await getKeywords()
        availableKeywords.value = res.data
      } catch (error) {
        console.error('加载关键字失败:', error)
      }
    }

    const editConfig = (row) => {
      editingId.value = row.id
      form.value = {
        web_page_id: row.web_page_id,
        crawl_time: row.crawl_time,
        crawl_freq: row.crawl_freq,
        keywords: row.keywords || []
      }
      
      if (row.crawl_time) {
        const timeStr = String(row.crawl_time).padStart(2, '0')
        crawlTime.value = `${timeStr}:00`
      } else {
        crawlTime.value = null
      }
      showDialog.value = true
    }

    const saveConfig = async () => {
      if (!form.value.web_page_id || !crawlTime.value || !form.value.crawl_freq || form.value.keywords.length === 0) {
        ElMessage.warning('请填写完整信息')
        return
      }

      if (crawlTime.value) {
        const timeStr = String(crawlTime.value)
        const hour = timeStr.split(':')[0]
        form.value.crawl_time = hour
      }

      try {
        if (editingId.value) {
          await updateMonitorConfig(editingId.value, form.value)
          ElMessage.success('更新成功')
        } else {
          await createMonitorConfig(form.value)
          ElMessage.success('添加成功')
        }
        showDialog.value = false
        editingId.value = null
        form.value = { web_page_id: null, crawl_time: '', crawl_freq: '', keywords: [] }
        crawlTime.value = null
        loadConfigs()
      } catch (error) {
        console.error('操作失败:', error)
        const message = error.response?.data?.error || error.message || '操作失败'
        ElMessage.error(message)
      }
    }

    const handleAdd = () => {
      editingId.value = null
      form.value = { web_page_id: null, crawl_time: '', crawl_freq: '', keywords: [] }
      crawlTime.value = null
      showDialog.value = true
    }

    const deleteConfig = async (id) => {
      try {
        await deleteMonitorConfig(id)
        ElMessage.success('删除成功')
        loadConfigs()
      } catch (error) {
        console.error('删除失败:', error)
        ElMessage.error('删除失败')
      }
    }

    onMounted(() => {
      loadConfigs()
      loadWebPages()
      loadKeywords()
    })

    return {
      configs,
      webPages,
      availableKeywords,
      showDialog,
      editingId,
      loading,
      crawlTime,
      form,
      editConfig,
      saveConfig,
      deleteConfig,
      handleAdd
    }
  }
}
</script>
