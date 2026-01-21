<template>
  <div>
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>采购信息动态</span>
          <div style="display: flex; gap: 10px">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索标题或内容"
              style="width: 300px"
              clearable
              @clear="handleSearch"
              @keyup.enter="handleSearch"
            />
            <el-button @click="handleSearch">搜索</el-button>
            <el-select v-model="sortOrder" @change="loadAnnouncements" style="width: 120px">
              <el-option label="最新优先" value="desc" />
              <el-option label="最早优先" value="asc" />
            </el-select>
          </div>
        </div>
      </template>

      <el-table :data="announcements" border v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="标题" min-width="250">
          <template #default="scope">
            <a :href="scope.row.url" target="_blank" style="color: #409eff; text-decoration: none">
              {{ scope.row.title }}
            </a>
          </template>
        </el-table-column>
        <el-table-column label="内容摘要" min-width="250">
          <template #default="scope">
            <div style="max-height: 60px; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; line-height: 1.6; color: #606266;">
              {{ getSummary(scope.row.content || scope.row.title) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="publisher" label="采购单位" width="180" />
        <el-table-column prop="web_page_name" label="来源" width="120" />
        <el-table-column prop="publish_date" label="发布时间" width="120" />
        <el-table-column prop="created_at" label="同步时间" width="180" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <el-button size="small" type="primary" link @click="showDetail(scope.row)">查看详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog v-model="detailVisible" title="公告详情" width="800px">
        <div v-if="currentDetail">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="标题" :span="2">
              <a :href="currentDetail.url" target="_blank" style="color: #409eff; text-decoration: none">
                {{ currentDetail.title }}
              </a>
            </el-descriptions-item>
            <el-descriptions-item label="采购单位">{{ currentDetail.publisher || '-' }}</el-descriptions-item>
            <el-descriptions-item label="来源">{{ currentDetail.web_page_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="发布时间">{{ currentDetail.publish_date }}</el-descriptions-item>
            <el-descriptions-item label="同步时间" :span="2">{{ currentDetail.created_at }}</el-descriptions-item>
          </el-descriptions>

          <el-divider>内容总结</el-divider>
          <div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 20px; border-radius: 8px; margin-bottom: 20px; color: white">
            <div style="font-size: 16px; font-weight: bold; margin-bottom: 15px">
              📋 关键信息摘要
            </div>
            <div style="background: rgba(255, 255, 255, 0.95); padding: 15px; border-radius: 6px; color: #303133; line-height: 1.8">
              <div style="margin-bottom: 12px">
                <div style="color: #909399; font-size: 12px; margin-bottom: 4px">标题</div>
                <div style="font-weight: 500">{{ currentDetail.title }}</div>
              </div>
              <div v-if="currentDetail.content && currentDetail.content.trim()" style="margin-bottom: 12px">
                <div style="color: #909399; font-size: 12px; margin-bottom: 4px">内容</div>
                <div>{{ currentDetail.content }}</div>
              </div>
              <div style="display: flex; gap: 20px; margin-top: 15px; padding-top: 15px; border-top: 1px solid #e4e7ed">
                <div>
                  <div style="color: #909399; font-size: 12px; margin-bottom: 4px">采购单位</div>
                  <div style="font-weight: 500">{{ currentDetail.publisher || '未知' }}</div>
                </div>
                <div>
                  <div style="color: #909399; font-size: 12px; margin-bottom: 4px">来源</div>
                  <div style="font-weight: 500">{{ currentDetail.web_page_name || '未知' }}</div>
                </div>
                <div>
                  <div style="color: #909399; font-size: 12px; margin-bottom: 4px">发布时间</div>
                  <div style="font-weight: 500">{{ currentDetail.publish_date }}</div>
                </div>
              </div>
            </div>
          </div>

          <el-divider>完整信息</el-divider>
          <div style="line-height: 1.8; color: #606266">
            <p><strong>标题：</strong>{{ currentDetail.title }}</p>
            <p><strong>采购单位：</strong>{{ currentDetail.publisher || '-' }}</p>
            <p><strong>链接：</strong><a :href="currentDetail.url" target="_blank" style="color: #409eff">{{ currentDetail.url }}</a></p>
            <p><strong>发布时间：</strong>{{ currentDetail.publish_date }}</p>
            <p><strong>同步时间：</strong>{{ currentDetail.created_at }}</p>
            <p v-if="currentDetail.content && currentDetail.content.trim()"><strong>内容：</strong>{{ currentDetail.content }}</p>
          </div>
        </div>
      </el-dialog>

      <div style="margin-top: 20px; display: flex; justify-content: center">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAnnouncements } from '../api'

export default {
  name: 'Announcements',
  setup() {
    const announcements = ref([])
    const loading = ref(false)
    const searchKeyword = ref('')
    const sortOrder = ref('desc')
    const currentPage = ref(1)
    const pageSize = ref(20)
    const total = ref(0)
    const detailVisible = ref(false)
    const currentDetail = ref(null)

    const getSummary = (text) => {
      if (!text) return '暂无内容'
      const maxLength = 120
      const trimmed = text.trim()
      if (trimmed.length <= maxLength) return trimmed
      const summary = trimmed.substring(0, maxLength)
      const lastSpace = summary.lastIndexOf(' ')
      if (lastSpace > maxLength - 30) {
        return summary.substring(0, lastSpace) + '...'
      }
      return summary + '...'
    }

    const showDetail = (row) => {
      currentDetail.value = row
      detailVisible.value = true
    }

    const loadAnnouncements = async () => {
      loading.value = true
      try {
        const res = await getAnnouncements({
          keyword: searchKeyword.value,
          order: sortOrder.value,
          page: currentPage.value,
          pageSize: pageSize.value
        })
        announcements.value = res.data.data || []
        total.value = res.data.total || 0
      } catch (error) {
        console.error('加载失败:', error)
        ElMessage.error('加载失败')
      } finally {
        loading.value = false
      }
    }

    const handleSearch = () => {
      currentPage.value = 1
      loadAnnouncements()
    }

    const handleSizeChange = () => {
      currentPage.value = 1
      loadAnnouncements()
    }

    const handlePageChange = () => {
      loadAnnouncements()
    }

    onMounted(() => {
      loadAnnouncements()
    })

    return {
      announcements,
      loading,
      searchKeyword,
      sortOrder,
      currentPage,
      pageSize,
      total,
      detailVisible,
      currentDetail,
      getSummary,
      showDetail,
      loadAnnouncements,
      handleSearch,
      handleSizeChange,
      handlePageChange
    }
  }
}
</script>
