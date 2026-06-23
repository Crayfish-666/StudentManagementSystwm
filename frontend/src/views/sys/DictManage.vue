<template>
  <div class="dict-manage">
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; align-items: center; justify-content: space-between">
          <span style="font-size: 16px; font-weight: 600">字典管理</span>
          <el-button type="primary" size="small" @click="showAddDialog">新增字典项</el-button>
        </div>
      </template>

      <el-tabs v-model="activeCategory" @tab-change="handleTabChange">
        <el-tab-pane
          v-for="cat in categories"
          :key="cat.category"
          :label="`${categoryNames[cat.category] || cat.category} (${cat.count})`"
          :name="cat.category"
        />
      </el-tabs>

      <el-table :data="currentItems" border stripe style="width: 100%">
        <el-table-column prop="id" label="编号" width="90" />
        <el-table-column prop="category" label="分类" width="140" />
        <el-table-column prop="code" label="编码" width="140" />
        <el-table-column prop="name_zh" label="中文名" width="160" />
        <el-table-column prop="name_en" label="英文名" width="140" />
        <el-table-column prop="sort" label="排序" width="70" />
        <el-table-column prop="is_active" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.is_active === 1 ? 'success' : 'info'" size="small">
              {{ row.is_active === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" text size="small" @click="showEditDialog(row)">编辑</el-button>
            <el-popconfirm title="确认删除？" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button type="danger" text size="small">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑字典项' : '新增字典项'" width="500px">
      <el-form :model="formData" label-width="80px">
        <el-form-item label="分类" required>
          <el-input v-model="formData.category" :disabled="isEdit" placeholder="如：gender" />
        </el-form-item>
        <el-form-item label="编码" required>
          <el-input v-model="formData.code" :disabled="isEdit" placeholder="如：M" />
        </el-form-item>
        <el-form-item label="中文名" required>
          <el-input v-model="formData.name_zh" placeholder="如：男" />
        </el-form-item>
        <el-form-item label="英文名">
          <el-input v-model="formData.name_en" placeholder="如：Male" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="formData.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { dictApi } from '@/api/sys'
import { useDictStore } from '@/stores/dict'

const dictStore = useDictStore()

const categories = ref([])
const activeCategory = ref('')
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)

const formData = ref({
  category: '',
  code: '',
  name_zh: '',
  name_en: '',
  sort: 0
})

// 分类中文名映射
const categoryNames = {
  gender: '性别',
  political_status: '政治面貌',
  activity_level: '活动等级',
  position_type: '岗位类型',
  difficulty_level: '困难等级',
  assoc_status: '社团状态',
  ethnicity: '民族',
  student_status: '学生状态',
  inspection_type: '巡查类型',
  incident_level: '事件等级'
}

// 当前分类下的字典项
const currentItems = computed(() => {
  const cat = categories.value.find(c => c.category === activeCategory.value)
  return cat?.items || []
})

onMounted(async () => {
  await loadCategories()
})

async function loadCategories() {
  try {
    const data = await dictApi.listCategories()
    categories.value = data.categories || []
    if (categories.value.length && !activeCategory.value) {
      activeCategory.value = categories.value[0].category
    }
  } catch (err) {
    ElMessage.error('加载字典分类失败')
  }
}

function handleTabChange() {
  // 切换 Tab 时数据已在前端
}

function showAddDialog() {
  isEdit.value = false
  editId.value = null
  formData.value = {
    category: activeCategory.value,
    code: '',
    name_zh: '',
    name_en: '',
    sort: 0
  }
  dialogVisible.value = true
}

function showEditDialog(row) {
  isEdit.value = true
  editId.value = row.id
  formData.value = {
    category: row.category,
    code: row.code,
    name_zh: row.name_zh,
    name_en: row.name_en || '',
    sort: row.sort
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formData.value.category || !formData.value.code || !formData.value.name_zh) {
    ElMessage.warning('请填写必填项')
    return
  }

  try {
    if (isEdit.value) {
      await dictApi.updateItem(editId.value, formData.value)
      ElMessage.success('修改成功')
    } else {
      await dictApi.createItem(formData.value)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    dictStore.clearCache(formData.value.category)
    await loadCategories()
  } catch (err) {
    ElMessage.error(isEdit.value ? '修改失败' : '新增失败')
  }
}

async function handleDelete(id) {
  try {
    await dictApi.deleteItem(id)
    ElMessage.success('删除成功')
    dictStore.clearCache()
    await loadCategories()
  } catch (err) {
    ElMessage.error('删除失败')
  }
}
</script>

<style scoped>
.dict-manage {
  padding: 0;
}
</style>
