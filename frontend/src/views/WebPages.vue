<template>
  <div>
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>网页列表管理</span>
          <el-button type="primary" @click="showDialog = true">添加网页</el-button>
        </div>
      </template>

      <el-table :data="webPages" border v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="url" label="URL" />
        <el-table-column label="操作" width="180">
          <template #default="scope">
            <el-button size="small" @click="editPage(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deletePage(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑网页' : '添加网页'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="URL">
          <el-input v-model="form.url" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="savePage">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getWebPages, createWebPage, updateWebPage, deleteWebPage } from '../api'

export default {
  name: 'WebPages',
  setup() {
    const webPages = ref([])
    const showDialog = ref(false)
    const editingId = ref(null)
    const loading = ref(false)
    const form = ref({ name: '', url: '' })

    const loadPages = async () => {
      loading.value = true
      try {
        const res = await getWebPages()
        webPages.value = res.data
      } catch (error) {
        console.error('加载失败:', error)
        const message = error.response?.data?.error || error.message || '加载失败'
        ElMessage.error(message)
      } finally {
        loading.value = false
      }
    }

    const editPage = (row) => {
      editingId.value = row.id
      form.value = { name: row.name, url: row.url }
      showDialog.value = true
    }

    const savePage = async () => {
      if (!form.value.name || !form.value.url) {
        ElMessage.warning('请填写完整信息')
        return
      }
      
      try {
        if (editingId.value) {
          await updateWebPage(editingId.value, form.value)
          ElMessage.success('更新成功')
        } else {
          await createWebPage(form.value)
          ElMessage.success('添加成功')
        }
        showDialog.value = false
        editingId.value = null
        form.value = { name: '', url: '' }
        loadPages()
      } catch (error) {
        console.error('操作失败:', error)
        const message = error.response?.data?.error || error.message || '操作失败'
        ElMessage.error(message)
      }
    }

    const deletePage = async (id) => {
      try {
        await deleteWebPage(id)
        ElMessage.success('删除成功')
        loadPages()
      } catch (error) {
        ElMessage.error('删除失败')
      }
    }

    onMounted(() => {
      loadPages()
    })

    return {
      webPages,
      showDialog,
      editingId,
      loading,
      form,
      editPage,
      savePage,
      deletePage
    }
  }
}
</script>
