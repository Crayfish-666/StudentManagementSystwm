<template>
  <div class="user-manage">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="card-title">用户管理</span>
          <div class="card-actions">
            <el-button type="primary" :icon="Plus" size="small" @click="showCreateDialog">新建用户</el-button>
            <el-button :icon="Refresh" size="small" @click="fetchList">刷新</el-button>
          </div>
        </div>
      </template>

      <!-- 筛选区 -->
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="关键词">
          <el-input
            v-model="filterForm.keyword"
            placeholder="账号 / 姓名 / 工号"
            clearable
            style="width: 220px"
            @keyup.enter="handleSearch"
            @clear="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filterForm.status" placeholder="全部" clearable style="width: 130px" @change="handleSearch">
            <el-option label="正常" value="active" />
            <el-option label="锁定" value="locked" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="filterForm.role_code" placeholder="全部" clearable filterable style="width: 200px" @change="handleSearch">
            <el-option
              v-for="r in roleOptions"
              :key="r.code"
              :label="`${r.name} (${r.code})`"
              :value="r.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">查询</el-button>
          <el-button :icon="RefreshLeft" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 列表 -->
      <el-table :data="list" stripe v-loading="loading" border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="账号" min-width="120" />
        <el-table-column prop="display_name" label="姓名" min-width="120" />
        <el-table-column prop="staff_no" label="工号" min-width="100">
          <template #default="{ row }">
            <span v-if="row.staff_no">{{ row.staff_no }}</span>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="角色" min-width="220">
          <template #default="{ row }">
            <el-tag
              v-for="role in row.roles"
              :key="role.id"
              :type="roleTagType(role.scope)"
              size="small"
              effect="light"
              style="margin-right: 4px; margin-bottom: 4px"
            >
              {{ role.name }}
            </el-tag>
            <span v-if="!row.roles || row.roles.length === 0" class="muted">未分配</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="dark">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_login_at" label="最近登录" width="170">
          <template #default="{ row }">
            <span v-if="row.last_login_at">{{ formatDateTime(row.last_login_at) }}</span>
            <span v-else class="muted">从未</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-button link type="primary" size="small" @click="showEditDialog(row)">编辑</el-button>
              <el-button link type="primary" size="small" @click="showAssignRolesDialog(row)">分配角色</el-button>
              <el-button link type="warning" size="small" @click="showResetPasswordDialog(row)">重置密码</el-button>
              <el-dropdown trigger="click" size="small" @command="(cmd) => handleStatusCmd(cmd, row)">
                <el-button link type="primary" size="small">更多</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-if="row.status !== 'locked'" :disabled="isSelf(row)" command="lock">锁定账户</el-dropdown-item>
                  <el-dropdown-item v-if="row.status === 'locked'" command="unlock">解锁账户</el-dropdown-item>
                  <el-dropdown-item v-if="row.status === 'active'" :disabled="isSelf(row)" command="disable" divided>禁用账户</el-dropdown-item>
                  <el-dropdown-item v-if="row.status === 'disabled'" command="enable">启用账户</el-dropdown-item>
                  <el-dropdown-item :disabled="isSelf(row)" command="delete" divided>
                    <span style="color: #f56c6c">删除用户</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="filterForm.page"
          v-model:page-size="filterForm.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="handleSearch"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 新建 / 编辑 对话框 -->
    <el-dialog
      v-model="formDialog.visible"
      :title="formDialog.isEdit ? '编辑用户' : '新建用户'"
      width="560px"
      destroy-on-close
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="90px"
        label-position="right"
      >
        <el-form-item label="账号" prop="username">
          <el-input v-model="form.username" :disabled="formDialog.isEdit" placeholder="学号 / 工号" />
        </el-form-item>
        <el-form-item v-if="!formDialog.isEdit" label="初始密码" prop="password">
          <el-input v-model="form.password" type="password" show-password placeholder="至少 6 位" />
        </el-form-item>
        <el-form-item label="姓名" prop="display_name">
          <el-input v-model="form.display_name" placeholder="用户姓名" />
        </el-form-item>
        <el-form-item label="工号">
          <el-input v-model="form.staff_no" placeholder="可选：教职工工号" />
        </el-form-item>
        <el-form-item v-if="!formDialog.isEdit" label="角色">
          <el-select v-model="form.role_ids" multiple placeholder="可选：分配角色" style="width: 100%">
            <el-option
              v-for="r in roleOptions"
              :key="r.id"
              :label="`${r.name} (${r.code})`"
              :value="r.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="formDialog.submitting" @click="handleFormSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 重置密码 对话框 -->
    <el-dialog v-model="passwordDialog.visible" title="重置密码" width="420px" destroy-on-close>
      <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-width="100px">
        <el-form-item label="目标用户">
          <span class="dialog-target">{{ passwordDialog.username }} ({{ passwordDialog.display_name }})</span>
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input v-model="passwordForm.new_password" type="password" show-password placeholder="至少 6 位" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm">
          <el-input v-model="passwordForm.confirm" type="password" show-password placeholder="再次输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="passwordDialog.submitting" @click="handleResetPasswordSubmit">提交</el-button>
      </template>
    </el-dialog>

    <!-- 分配角色 对话框 -->
    <el-dialog
      v-model="roleDialog.visible"
      title="分配角色"
      width="520px"
      destroy-on-close
      @close="resetRoleDialog"
    >
      <div class="role-target">
        目标用户：<b>{{ roleDialog.username }}</b> ({{ roleDialog.display_name }})
      </div>
      <el-form label-width="90px">
        <el-form-item label="当前角色">
          <el-tag
            v-for="r in roleDialog.currentRoles"
            :key="r.id"
            closable
            :type="roleTagType(r.scope)"
            style="margin-right: 4px; margin-bottom: 4px"
            @close="handleRemoveRole(r)"
          >
            {{ r.name }}
          </el-tag>
          <span v-if="roleDialog.currentRoles.length === 0" class="muted">暂无角色</span>
        </el-form-item>
        <el-form-item label="新增角色">
          <el-select
            v-model="roleDialog.pendingAdd"
            placeholder="选择角色并点击「添加」"
            filterable
            style="width: 70%"
          >
            <el-option
              v-for="r in availableRoles"
              :key="r.id"
              :label="`${r.name} (${r.code})`"
              :value="r.id"
            />
          </el-select>
          <el-button type="primary" :disabled="!roleDialog.pendingAdd" @click="handleAddRole" style="margin-left: 8px">
            添加
          </el-button>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="roleDialog.submitting" @click="handleSaveRoles">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { userApi, roleApi } from '@/api/sys'
import { useAuthStore } from '@/stores/auth'
import { formatDateTime } from '@/utils/datetime'

// 当前登录用户
const authStore = useAuthStore()
const isSelf = (row) => authStore.user && row.id === authStore.user.id

// 状态相关
const loading = ref(false)
const list = ref([])
const total = ref(0)
const roleOptions = ref([])

// 筛选
const filterForm = reactive({
  keyword: '',
  status: '',
  role_code: '',
  page: 1,
  page_size: 20
})

// 表单（新建/编辑）
const formRef = ref(null)
const formDialog = reactive({ visible: false, isEdit: false, submitting: false, editId: null })
const form = reactive({
  username: '',
  password: '',
  display_name: '',
  staff_no: '',
  role_ids: []
})
const formRules = {
  username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, min: 6, message: '密码长度至少 6 位', trigger: 'blur' }],
  display_name: [{ required: true, message: '请输入姓名', trigger: 'blur' }]
}

// 重置密码
const passwordFormRef = ref(null)
const passwordDialog = reactive({ visible: false, submitting: false, userId: null, username: '', display_name: '' })
const passwordForm = reactive({ new_password: '', confirm: '' })
const passwordRules = {
  new_password: [{ required: true, min: 6, message: '新密码长度至少 6 位', trigger: 'blur' }],
  confirm: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (_rule, value, cb) => {
        if (value !== passwordForm.new_password) cb(new Error('两次输入的密码不一致'))
        else cb()
      },
      trigger: 'blur'
    }
  ]
}

// 分配角色
const roleDialog = reactive({
  visible: false,
  submitting: false,
  userId: null,
  username: '',
  display_name: '',
  currentRoles: [],
  pendingAdd: null
})

// 计算未分配的可选角色
const availableRoles = computed(() => {
  const ids = new Set(roleDialog.currentRoles.map((r) => r.id))
  return roleOptions.value.filter((r) => !ids.has(r.id))
})

// ============ 数据加载 ============

const fetchList = async () => {
  loading.value = true
  try {
    const data = await userApi.list({
      keyword: filterForm.keyword || undefined,
      status: filterForm.status || undefined,
      role_code: filterForm.role_code || undefined,
      page: filterForm.page,
      page_size: filterForm.page_size
    })
    list.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    // http 拦截器已提示
  } finally {
    loading.value = false
  }
}

const fetchRoles = async () => {
  try {
    const data = await roleApi.list()
    roleOptions.value = data.items || []
  } catch (e) {
    roleOptions.value = []
  }
}

const handleSearch = () => {
  filterForm.page = 1
  fetchList()
}

const handleReset = () => {
  filterForm.keyword = ''
  filterForm.status = ''
  filterForm.role_code = ''
  filterForm.page = 1
  fetchList()
}

// ============ 新建 / 编辑 ============

const resetForm = () => {
  form.username = ''
  form.password = ''
  form.display_name = ''
  form.staff_no = ''
  form.role_ids = []
  formRef.value?.clearValidate()
}

const showCreateDialog = () => {
  formDialog.isEdit = false
  formDialog.editId = null
  formDialog.visible = true
  nextTick(resetForm)
}

const showEditDialog = (row) => {
  formDialog.isEdit = true
  formDialog.editId = row.id
  form.username = row.username
  form.password = ''
  form.display_name = row.display_name
  form.staff_no = row.staff_no || ''
  formDialog.visible = true
}

const handleFormSubmit = async () => {
  await formRef.value.validate()
  formDialog.submitting = true
  try {
    if (formDialog.isEdit) {
      await userApi.update(formDialog.editId, {
        display_name: form.display_name,
        staff_no: form.staff_no
      })
      ElMessage.success('已保存')
    } else {
      await userApi.create({
        username: form.username,
        password: form.password,
        display_name: form.display_name,
        staff_no: form.staff_no,
        role_ids: form.role_ids
      })
      ElMessage.success('已创建')
    }
    formDialog.visible = false
    fetchList()
  } catch (e) {
    // ignore
  } finally {
    formDialog.submitting = false
  }
}

// ============ 重置密码 ============

const showResetPasswordDialog = (row) => {
  passwordDialog.userId = row.id
  passwordDialog.username = row.username
  passwordDialog.display_name = row.display_name
  passwordForm.new_password = ''
  passwordForm.confirm = ''
  passwordDialog.visible = true
  nextTick(() => passwordFormRef.value?.clearValidate())
}

const handleResetPasswordSubmit = async () => {
  await passwordFormRef.value.validate()
  try {
    await ElMessageBox.confirm(
      `确定将用户 [${passwordDialog.username}] 的密码重置吗？`,
      '重置密码',
      { type: 'warning' }
    )
  } catch {
    return
  }
  passwordDialog.submitting = true
  try {
    await userApi.resetPassword(passwordDialog.userId, passwordForm.new_password)
    ElMessage.success('密码已重置')
    passwordDialog.visible = false
  } catch (e) {
    // ignore
  } finally {
    passwordDialog.submitting = false
  }
}

// ============ 分配角色 ============

const showAssignRolesDialog = (row) => {
  roleDialog.userId = row.id
  roleDialog.username = row.username
  roleDialog.display_name = row.display_name
  roleDialog.currentRoles = (row.roles || []).map((r) => ({ ...r }))
  roleDialog.pendingAdd = null
  roleDialog.visible = true
}

const resetRoleDialog = () => {
  roleDialog.userId = null
  roleDialog.username = ''
  roleDialog.display_name = ''
  roleDialog.currentRoles = []
  roleDialog.pendingAdd = null
}

const handleAddRole = () => {
  if (!roleDialog.pendingAdd) return
  const role = roleOptions.value.find((r) => r.id === roleDialog.pendingAdd)
  if (role && !roleDialog.currentRoles.find((r) => r.id === role.id)) {
    roleDialog.currentRoles.push({ ...role })
  }
  roleDialog.pendingAdd = null
}

const handleRemoveRole = (role) => {
  roleDialog.currentRoles = roleDialog.currentRoles.filter((r) => r.id !== role.id)
}

const handleSaveRoles = async () => {
  roleDialog.submitting = true
  try {
    const roleIds = roleDialog.currentRoles.map((r) => r.id)
    await userApi.assignRoles(roleDialog.userId, roleIds)
    ElMessage.success('已保存')
    roleDialog.visible = false
    fetchList()
  } catch (e) {
    // ignore
  } finally {
    roleDialog.submitting = false
  }
}

// ============ 状态切换 / 删除 ============

const handleStatusCmd = async (cmd, row) => {
  if (isSelf(row) && ['lock', 'disable', 'delete'].includes(cmd)) {
    ElMessage.warning('不能对自己执行此操作')
    return
  }
  const map = {
    lock: { msg: '锁定后该用户将无法登录，确认吗？', action: () => userApi.lock(row.id) },
    unlock: { msg: '确认解锁该用户？', action: () => userApi.unlock(row.id) },
    disable: { msg: '禁用后该用户将无法登录，确认吗？', action: () => userApi.disable(row.id) },
    enable: { msg: '确认启用该用户？', action: () => userApi.enable(row.id) },
    delete: { msg: `确认删除用户 [${row.username}] 吗？删除后无法恢复`, action: () => userApi.remove(row.id) }
  }
  const item = map[cmd]
  if (!item) return
  try {
    await ElMessageBox.confirm(item.msg, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await item.action()
    ElMessage.success('操作成功')
    fetchList()
  } catch (e) {
    // ignore
  }
}

// ============ 工具函数 ============

const statusText = (s) => ({ active: '正常', locked: '锁定', disabled: '禁用' }[s] || s)
const statusTagType = (s) => ({ active: 'success', locked: 'warning', disabled: 'info' }[s] || 'info')
const roleTagType = (scope) => ({ school: 'danger', college: 'warning', student: 'success' }[scope] || 'info')

onMounted(async () => {
  await fetchRoles()
  await fetchList()
})
</script>

<style scoped>
.user-manage {
  padding: var(--sh-space-xs);
}
/* .card-header 已在 App.vue 全局定义 */
.action-btns {
  display: flex;
  align-items: center;
  gap: var(--sh-space-md);
}
.action-btns :deep(.el-button),
.action-btns :deep(.el-dropdown) {
  margin: 0 !important;
}
.card-title {
  font-size: var(--sh-text-lg);
  font-weight: 600;
  color: var(--sh-text-primary);
}
.filter-form {
  margin-bottom: var(--sh-space-sm);
}
.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--sh-space-md);
}
.muted {
  color: var(--sh-text-disabled);
  font-size: var(--sh-text-xs);
}
.dialog-target {
  color: var(--sh-text-primary);
  font-weight: 500;
}
.role-target {
  margin-bottom: var(--sh-space-sm);
  color: var(--sh-text-regular);
}
</style>
