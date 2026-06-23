<template>
  <div class="building-tree-page">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <el-button type="primary" @click="handleAddBuilding">
        <el-icon><Plus /></el-icon>
        新增楼栋
      </el-button>
    </div>

    <div class="main-content">
      <!-- 左侧树形结构 -->
      <div class="tree-panel">
        <el-tree
          ref="treeRef"
          :data="treeData"
          :props="treeProps"
          node-key="compositeKey"
          highlight-current
          default-expand-all
          @node-click="handleNodeClick"
        >
          <template #default="{ node, data }">
            <span class="tree-node-label">
              <el-icon class="tree-icon"><component :is="getNodeIcon(data.type)" /></el-icon>
              <span>{{ data.name }}</span>
            </span>
          </template>
        </el-tree>
      </div>

      <!-- 右侧详情/编辑面板 -->
      <div class="detail-panel">
        <template v-if="!selectedNode">
          <el-empty description="请在左侧选择节点查看详情" />
        </template>

        <!-- 楼栋详情 -->
        <template v-else-if="selectedNode.type === 'building'">
          <div class="detail-header">
            <h3>楼栋信息</h3>
            <div class="detail-actions">
              <el-button type="primary" size="small" @click="handleEditBuilding(selectedNode)">编辑</el-button>
              <el-popconfirm title="确定删除该楼栋吗？" @confirm="handleDeleteBuilding(selectedNode)">
                <template #reference>
                  <el-button type="danger" size="small">删除</el-button>
                </template>
              </el-popconfirm>
            </div>
          </div>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="编码">{{ selectedNode.code }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ selectedNode.name }}</el-descriptions-item>
            <el-descriptions-item label="楼层数">{{ selectedNode.floor_count ?? (selectedNode.children?.length ?? 0) }}</el-descriptions-item>
          </el-descriptions>
        </template>

        <!-- 楼层详情 -->
        <template v-else-if="selectedNode.type === 'floor'">
          <div class="detail-header">
            <h3>楼层信息</h3>
          </div>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="楼层号">{{ selectedNode.floor_no }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ selectedNode.name }}</el-descriptions-item>
          </el-descriptions>
          <h4 style="margin-top: 20px;">寝室列表</h4>
          <el-table :data="selectedNode.children || []" border stripe>
            <el-table-column prop="room_no" label="寝室号" />
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="bed_count" label="床位数" />
          </el-table>
        </template>

        <!-- 寝室详情 -->
        <template v-else-if="selectedNode.type === 'room'">
          <div class="detail-header">
            <h3>寝室信息</h3>
          </div>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="寝室号">{{ selectedNode.room_no }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ selectedNode.name }}</el-descriptions-item>
            <el-descriptions-item label="床位数">{{ selectedNode.bed_count }}</el-descriptions-item>
          </el-descriptions>
          <h4 style="margin-top: 20px;">入住成员</h4>
          <el-table :data="selectedNode.members || []" border stripe>
            <el-table-column prop="name" label="姓名" />
            <el-table-column prop="student_no" label="学号" />
          </el-table>
          <el-empty v-if="!selectedNode.members?.length" description="暂无入住成员" />
        </template>
      </div>
    </div>

    <!-- 新增/编辑楼栋弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'add' ? '新增楼栋' : '编辑楼栋'"
      width="480px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-form-item label="编码" prop="code">
          <el-input v-model="formData.code" placeholder="请输入楼栋编码" />
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入楼栋名称" />
        </el-form-item>
        <el-form-item label="楼层数" prop="floor_count">
          <el-input-number v-model="formData.floor_count" :min="0" />
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { House, OfficeBuilding, Key, Plus } from '@element-plus/icons-vue'
import { sqBuildingApi } from '@/api/sq'

// ========== 树相关 ==========

const treeRef = ref(null)
const treeData = ref([])
const treeProps = {
  children: 'children',
  label: 'name',
}

// 为每个节点添加 compositeKey，避免不同层级 id 冲突
function addCompositeKey(nodes) {
  if (!nodes) return
  for (const node of nodes) {
    node.compositeKey = node.type + '-' + node.id
    if (node.children?.length) {
      addCompositeKey(node.children)
    }
  }
}

async function fetchTree() {
  try {
    const data = await sqBuildingApi.tree()
    const items = data?.items || []
    addCompositeKey(items)
    treeData.value = items
  } catch {
    ElMessage.error('获取楼栋树失败')
  }
}

function getNodeIcon(type) {
  switch (type) {
    case 'building': return House
    case 'floor': return OfficeBuilding
    case 'room': return Key
    default: return House
  }
}

// ========== 选中节点 ==========

const selectedNode = ref(null)

function handleNodeClick(data) {
  selectedNode.value = data
}

// ========== 弹窗相关 ==========

const dialogVisible = ref(false)
const dialogMode = ref('add') // 'add' | 'edit'
const formRef = ref(null)
const formData = reactive({
  id: null,
  code: '',
  name: '',
  floor_count: 0,
})

const formRules = {
  code: [{ required: true, message: '请输入楼栋编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入楼栋名称', trigger: 'blur' }],
  floor_count: [{ required: true, message: '请输入楼层数', trigger: 'blur' }],
}

function resetForm() {
  formData.id = null
  formData.code = ''
  formData.name = ''
  formData.floor_count = 0
}

function handleAddBuilding() {
  resetForm()
  dialogMode.value = 'add'
  dialogVisible.value = true
}

function handleEditBuilding(node) {
  formData.id = node.id
  formData.code = node.code
  formData.name = node.name
  formData.floor_count = node.floor_count ?? (node.children?.length ?? 0)
  dialogMode.value = 'edit'
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  try {
    if (dialogMode.value === 'add') {
      await sqBuildingApi.create(formData)
      ElMessage.success('新增楼栋成功')
    } else {
      await sqBuildingApi.update(formData.id, formData)
      ElMessage.success('编辑楼栋成功')
    }
    dialogVisible.value = false
    selectedNode.value = null
    await fetchTree()
  } catch {
    ElMessage.error(dialogMode.value === 'add' ? '新增楼栋失败' : '编辑楼栋失败')
  }
}

// ========== 删除 ==========

async function handleDeleteBuilding(node) {
  try {
    await sqBuildingApi.delete(node.id)
    ElMessage.success('删除楼栋成功')
    selectedNode.value = null
    await fetchTree()
  } catch {
    ElMessage.error('删除楼栋失败')
  }
}

// ========== 初始化 ==========

onMounted(() => {
  fetchTree()
})
</script>

<style scoped>
.building-tree-page {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.toolbar {
  padding: var(--sh-space-sm) var(--sh-space-md);
  border-bottom: 1px solid var(--sh-border-light);
  flex-shrink: 0;
}

.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.tree-panel {
  width: 300px;
  border-right: 1px solid var(--sh-border-light);
  overflow-y: auto;
  padding: var(--sh-space-sm);
  flex-shrink: 0;
}

.detail-panel {
  flex: 1;
  padding: var(--sh-space-lg);
  overflow-y: auto;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--sh-space-md);
}

.detail-header h3 {
  margin: 0;
  font-size: var(--sh-text-lg);
  color: var(--sh-text-primary);
}

.detail-actions {
  display: flex;
  gap: var(--sh-space-sm);
}

.tree-node-label {
  display: flex;
  align-items: center;
  gap: var(--sh-space-xs);
}

.tree-icon {
  font-size: var(--sh-text-lg);
}
</style>
