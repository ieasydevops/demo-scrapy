<template>
  <div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="关键字管理" name="keywords">
        <el-form :inline="true" style="margin-bottom: 20px">
          <el-form-item label="关键字">
            <el-input v-model="newKeyword" placeholder="输入关键字" style="width: 200px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="addKeyword">添加</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="keywords" border>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="keyword" label="关键字" />
          <el-table-column label="操作" width="120">
            <template #default="scope">
              <el-button size="small" type="danger" @click="deleteKeyword(scope.row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getKeywords, createKeyword, deleteKeyword as deleteKeywordApi } from '../api'

export default {
  name: 'Keywords',
  setup() {
    const keywords = ref([])
    const newKeyword = ref('')
    const activeTab = ref('keywords')

    const loadKeywords = async () => {
      try {
        const res = await getKeywords()
        keywords.value = res.data
      } catch (error) {
        console.error('加载失败:', error)
        const message = error.response?.data?.error || error.message || '加载失败'
        ElMessage.error(message)
      }
    }

    const addKeyword = async () => {
      if (!newKeyword.value.trim()) {
        ElMessage.warning('请输入关键字')
        return
      }
      try {
        await createKeyword({ keyword: newKeyword.value })
        ElMessage.success('添加成功')
        newKeyword.value = ''
        loadKeywords()
      } catch (error) {
        console.error('添加失败:', error)
        const message = error.response?.data?.error || error.message || '添加失败'
        ElMessage.error(message)
      }
    }

    const deleteKeyword = async (id) => {
      try {
        await deleteKeywordApi(id)
        ElMessage.success('删除成功')
        loadKeywords()
      } catch (error) {
        ElMessage.error('删除失败')
      }
    }

    onMounted(() => {
      loadKeywords()
    })

    return {
      keywords,
      newKeyword,
      activeTab,
      addKeyword,
      deleteKeyword
    }
  }
}
</script>
