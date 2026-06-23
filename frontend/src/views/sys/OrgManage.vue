<template>
  <div class="org-manage">
    <el-card shadow="hover">
      <template #header>
        <span style="font-size: 16px; font-weight: 600">组织管理</span>
      </template>

      <el-tabs v-model="activeTab">
        <!-- 院系管理 -->
        <el-tab-pane label="院系管理" name="college">
          <div style="margin-bottom: 12px">
            <el-button type="primary" size="small" @click="showCollegeForm()">新增院系</el-button>
          </div>
          <el-table :data="colleges" stripe v-loading="collegeLoading">
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="院系名称" min-width="200" />
            <el-table-column prop="code" label="院系编码" width="140" />
            <el-table-column prop="dean" label="院长" width="120" />
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="showCollegeForm(row)">编辑</el-button>
                <el-popconfirm title="确认删除？删除前需确保该院系下无专业" @confirm="handleDeleteCollege(row.id)">
                  <template #reference>
                    <el-button link type="danger" size="small">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 专业管理 -->
        <el-tab-pane label="专业管理" name="major">
          <div style="margin-bottom: 12px; display: flex; gap: 12px; align-items: center">
            <el-select v-model="majorFilterCollege" placeholder="按院系筛选" clearable style="width: 200px" @change="fetchMajors">
              <el-option v-for="c in colleges" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
            <el-button type="primary" size="small" @click="showMajorForm()">新增专业</el-button>
          </div>
          <el-table :data="majors" stripe v-loading="majorLoading">
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="专业名称" min-width="180" />
            <el-table-column prop="code" label="专业编码" width="140" />
            <el-table-column prop="college_id" label="所属院系ID" width="100">
              <template #default="{ row }">
                {{ getCollegeName(row.college_id) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="showMajorForm(row)">编辑</el-button>
                <el-popconfirm title="确认删除？删除前需确保该专业下无班级" @confirm="handleDeleteMajor(row.id)">
                  <template #reference>
                    <el-button link type="danger" size="small">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 班级管理 -->
        <el-tab-pane label="班级管理" name="class">
          <div style="margin-bottom: 12px; display: flex; gap: 12px; align-items: center">
            <el-select v-model="classFilterMajor" placeholder="按专业筛选" clearable style="width: 200px" @change="fetchClasses">
              <el-option v-for="m in majors" :key="m.id" :label="m.name" :value="m.id" />
            </el-select>
            <el-button type="primary" size="small" @click="showClassForm()">新增班级</el-button>
          </div>
          <el-table :data="classes" stripe v-loading="classLoading">
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="班级名称" min-width="180" />
            <el-table-column prop="code" label="班级编码" width="140" />
            <el-table-column prop="major_id" label="所属专业ID" width="100">
              <template #default="{ row }">
                {{ getMajorName(row.major_id) }}
              </template>
            </el-table-column>
            <el-table-column prop="grade" label="年级" width="80" />
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="showClassForm(row)">编辑</el-button>
                <el-popconfirm title="确认删除？删除前需确保该班级下无学生" @confirm="handleDeleteClass(row.id)">
                  <template #reference>
                    <el-button link type="danger" size="small">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 院系弹窗 -->
    <el-dialog v-model="collegeDialogVisible" :title="collegeIsEdit ? '编辑院系' : '新增院系'" width="480px" destroy-on-close>
      <el-form ref="collegeFormRef" :model="collegeForm" :rules="collegeRules" label-width="80px">
        <el-form-item label="院系名称" prop="name">
          <el-input v-model="collegeForm.name" />
        </el-form-item>
        <el-form-item label="院系编码" prop="code">
          <el-input v-model="collegeForm.code" />
        </el-form-item>
        <el-form-item label="院长">
          <el-input v-model="collegeForm.dean" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="collegeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="collegeSubmitting" @click="handleCollegeSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 专业弹窗 -->
    <el-dialog v-model="majorDialogVisible" :title="majorIsEdit ? '编辑专业' : '新增专业'" width="480px" destroy-on-close>
      <el-form ref="majorFormRef" :model="majorForm" :rules="majorRules" label-width="80px">
        <el-form-item label="专业名称" prop="name">
          <el-input v-model="majorForm.name" />
        </el-form-item>
        <el-form-item label="专业编码" prop="code">
          <el-input v-model="majorForm.code" />
        </el-form-item>
        <el-form-item label="所属院系" prop="college_id">
          <el-select v-model="majorForm.college_id" placeholder="请选择院系" style="width: 100%">
            <el-option v-for="c in colleges" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="majorDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="majorSubmitting" @click="handleMajorSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 班级弹窗 -->
    <el-dialog v-model="classDialogVisible" :title="classIsEdit ? '编辑班级' : '新增班级'" width="480px" destroy-on-close>
      <el-form ref="classFormRef" :model="classForm" :rules="classRules" label-width="80px">
        <el-form-item label="班级名称" prop="name">
          <el-input v-model="classForm.name" />
        </el-form-item>
        <el-form-item label="班级编码" prop="code">
          <el-input v-model="classForm.code" />
        </el-form-item>
        <el-form-item label="所属专业" prop="major_id">
          <el-select v-model="classForm.major_id" placeholder="请选择专业" style="width: 100%">
            <el-option v-for="m in majors" :key="m.id" :label="m.name" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="年级">
          <el-input-number v-model="classForm.grade" :min="2000" :max="2099" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="classDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="classSubmitting" @click="handleClassSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { collegeApi, majorApi, classApi } from '@/api/sys-org'

const activeTab = ref('college')

// ========== 院系 ==========
const colleges = ref([])
const collegeLoading = ref(false)
const collegeDialogVisible = ref(false)
const collegeIsEdit = ref(false)
const collegeEditId = ref(null)
const collegeSubmitting = ref(false)
const collegeFormRef = ref()
const collegeForm = ref({ name: '', code: '', dean: '' })
const collegeRules = {
  name: [{ required: true, message: '请输入院系名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入院系编码', trigger: 'blur' }]
}

async function fetchColleges() {
  collegeLoading.value = true
  try {
    const data = await collegeApi.list()
    colleges.value = data.items || []
  } catch (e) {
    console.error('获取院系列表失败', e)
  } finally {
    collegeLoading.value = false
  }
}

function showCollegeForm(row) {
  if (row) {
    collegeIsEdit.value = true
    collegeEditId.value = row.id
    collegeForm.value = { name: row.name, code: row.code, dean: row.dean || '' }
  } else {
    collegeIsEdit.value = false
    collegeEditId.value = null
    collegeForm.value = { name: '', code: '', dean: '' }
  }
  collegeDialogVisible.value = true
}

async function handleCollegeSubmit() {
  try {
    await collegeFormRef.value.validate()
  } catch { return }

  collegeSubmitting.value = true
  try {
    if (collegeIsEdit.value) {
      await collegeApi.update(collegeEditId.value, collegeForm.value)
      ElMessage.success('更新成功')
    } else {
      await collegeApi.create(collegeForm.value)
      ElMessage.success('创建成功')
    }
    collegeDialogVisible.value = false
    fetchColleges()
  } catch (e) { /* 错误已由拦截器处理 */ } finally {
    collegeSubmitting.value = false
  }
}

async function handleDeleteCollege(id) {
  try {
    await collegeApi.delete(id)
    ElMessage.success('删除成功')
    fetchColleges()
  } catch (e) { /* 错误已由拦截器处理 */ }
}

// ========== 专业 ==========
const majors = ref([])
const majorLoading = ref(false)
const majorFilterCollege = ref(null)
const majorDialogVisible = ref(false)
const majorIsEdit = ref(false)
const majorEditId = ref(null)
const majorSubmitting = ref(false)
const majorFormRef = ref()
const majorForm = ref({ name: '', code: '', college_id: null })
const majorRules = {
  name: [{ required: true, message: '请输入专业名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入专业编码', trigger: 'blur' }],
  college_id: [{ required: true, message: '请选择所属院系', trigger: 'change' }]
}

async function fetchMajors() {
  majorLoading.value = true
  try {
    const params = {}
    if (majorFilterCollege.value) params.college_id = majorFilterCollege.value
    const data = await majorApi.list(params)
    majors.value = data.items || []
  } catch (e) {
    console.error('获取专业列表失败', e)
  } finally {
    majorLoading.value = false
  }
}

function showMajorForm(row) {
  if (row) {
    majorIsEdit.value = true
    majorEditId.value = row.id
    majorForm.value = { name: row.name, code: row.code, college_id: row.college_id }
  } else {
    majorIsEdit.value = false
    majorEditId.value = null
    majorForm.value = { name: '', code: '', college_id: null }
  }
  majorDialogVisible.value = true
}

async function handleMajorSubmit() {
  try {
    await majorFormRef.value.validate()
  } catch { return }

  majorSubmitting.value = true
  try {
    if (majorIsEdit.value) {
      await majorApi.update(majorEditId.value, majorForm.value)
      ElMessage.success('更新成功')
    } else {
      await majorApi.create(majorForm.value)
      ElMessage.success('创建成功')
    }
    majorDialogVisible.value = false
    fetchMajors()
  } catch (e) { /* 错误已由拦截器处理 */ } finally {
    majorSubmitting.value = false
  }
}

async function handleDeleteMajor(id) {
  try {
    await majorApi.delete(id)
    ElMessage.success('删除成功')
    fetchMajors()
  } catch (e) { /* 错误已由拦截器处理 */ }
}

// ========== 班级 ==========
const classes = ref([])
const classLoading = ref(false)
const classFilterMajor = ref(null)
const classDialogVisible = ref(false)
const classIsEdit = ref(false)
const classEditId = ref(null)
const classSubmitting = ref(false)
const classFormRef = ref()
const classForm = ref({ name: '', code: '', major_id: null, grade: new Date().getFullYear() })
const classRules = {
  name: [{ required: true, message: '请输入班级名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入班级编码', trigger: 'blur' }],
  major_id: [{ required: true, message: '请选择所属专业', trigger: 'change' }]
}

async function fetchClasses() {
  classLoading.value = true
  try {
    const params = {}
    if (classFilterMajor.value) params.major_id = classFilterMajor.value
    const data = await classApi.list(params)
    classes.value = data.items || []
  } catch (e) {
    console.error('获取班级列表失败', e)
  } finally {
    classLoading.value = false
  }
}

function showClassForm(row) {
  if (row) {
    classIsEdit.value = true
    classEditId.value = row.id
    classForm.value = { name: row.name, code: row.code, major_id: row.major_id, grade: row.grade }
  } else {
    classIsEdit.value = false
    classEditId.value = null
    classForm.value = { name: '', code: '', major_id: null, grade: new Date().getFullYear() }
  }
  classDialogVisible.value = true
}

async function handleClassSubmit() {
  try {
    await classFormRef.value.validate()
  } catch { return }

  classSubmitting.value = true
  try {
    if (classIsEdit.value) {
      await classApi.update(classEditId.value, classForm.value)
      ElMessage.success('更新成功')
    } else {
      await classApi.create(classForm.value)
      ElMessage.success('创建成功')
    }
    classDialogVisible.value = false
    fetchClasses()
  } catch (e) { /* 错误已由拦截器处理 */ } finally {
    classSubmitting.value = false
  }
}

async function handleDeleteClass(id) {
  try {
    await classApi.delete(id)
    ElMessage.success('删除成功')
    fetchClasses()
  } catch (e) { /* 错误已由拦截器处理 */ }
}

// ========== 辅助 ==========
function getCollegeName(collegeId) {
  const c = colleges.value.find(item => item.id === collegeId)
  return c ? c.name : collegeId
}

function getMajorName(majorId) {
  const m = majors.value.find(item => item.id === majorId)
  return m ? m.name : majorId
}

onMounted(() => {
  fetchColleges()
  fetchMajors()
  fetchClasses()
})
</script>

<style scoped>
.org-manage {
  padding: 0;
}
</style>
