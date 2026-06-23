<template>
  <div class="page-container">
    <h2>巡查录入</h2>

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      style="max-width: 700px"
    >
      <el-form-item label="巡查类型" prop="inspection_type">
        <el-select v-model="form.inspection_type" placeholder="请选择巡查类型">
          <el-option
            v-for="(label, key) in typeDict"
            :key="key"
            :label="label"
            :value="key"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="楼栋" prop="building_id">
        <el-select
          v-model="form.building_id"
          placeholder="请选择楼栋"
          @change="onBuildingChange"
        >
          <el-option
            v-for="b in buildings"
            :key="b.id"
            :label="b.name"
            :value="b.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="楼层" prop="floor_id">
        <el-select v-model="form.floor_id" placeholder="请选择楼层" clearable>
          <el-option
            v-for="f in floors"
            :key="f.id"
            :label="f.floor_no + ' 层'"
            :value="f.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="寝室" prop="room_id">
        <el-select v-model="form.room_id" placeholder="请选择寝室" clearable filterable>
          <el-option
            v-for="r in rooms"
            :key="r.id"
            :label="r.room_no"
            :value="r.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="巡查时间" prop="inspected_at">
        <el-date-picker
          v-model="form.inspected_at"
          type="datetime"
          placeholder="请选择巡查时间"
          value-format="YYYY-MM-DD HH:mm:ss"
        />
      </el-form-item>

      <el-form-item label="摘要" prop="summary">
        <el-input
          v-model="form.summary"
          type="textarea"
          :rows="3"
          placeholder="请输入巡查摘要"
        />
      </el-form-item>

      <!-- 扣分项 -->
      <el-form-item label="扣分项">
        <el-table :data="form.deductions" border style="width: 100%">
          <el-table-column label="扣分项" min-width="300">
            <template #default="{ row }">
              <el-input v-model="row.item" placeholder="扣分项名称" />
            </template>
          </el-table-column>
          <el-table-column label="扣分" width="140">
            <template #default="{ row }">
              <el-input-number v-model="row.deduction" :min="0" :max="100" :precision="0" controls-position="right" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" align="center">
            <template #default="{ $index }">
              <el-button link type="danger" @click="removeDeduction($index)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-button type="primary" link style="margin-top: 8px" @click="addDeduction">+ 添加扣分项</el-button>
      </el-form-item>

      <el-form-item label="预估分数">
        <span class="score-preview" :class="scoreClass">{{ computedScore }} 分</span>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">提交</el-button>
        <el-button @click="router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { sqInspectionApi, sqBuildingApi } from '@/api/sq'

const router = useRouter()

// 巡查类型字典
const typeDict = {
  hygiene: '卫生巡查',
  late_return: '晚归检查',
  appliance: '违规电器',
  safety: '安全隐患',
  fire_lane: '消防通道'
}

// 表单
const formRef = ref(null)
const submitting = ref(false)

const form = reactive({
  inspection_type: '',
  building_id: '',
  floor_id: '',
  room_id: '',
  inspected_at: '',
  summary: '',
  deductions: []
})

const rules = {
  inspection_type: [{ required: true, message: '请选择巡查类型', trigger: 'change' }],
  building_id: [{ required: true, message: '请选择楼栋', trigger: 'change' }],
  inspected_at: [{ required: true, message: '请选择巡查时间', trigger: 'change' }]
}

// 下拉数据
const buildings = ref([])
const floors = ref([])
const rooms = ref([])

// 计算分数
const computedScore = computed(() => {
  const totalDeduction = form.deductions.reduce((sum, d) => sum + (d.deduction || 0), 0)
  return Math.max(0, 100 - totalDeduction)
})

const scoreClass = computed(() => {
  if (computedScore.value >= 90) return 'score-green'
  if (computedScore.value >= 60) return 'score-orange'
  return 'score-red'
})

// 楼栋选择变化
async function onBuildingChange(buildingId) {
  // 清空楼层和寝室
  form.floor_id = ''
  form.room_id = ''
  floors.value = []
  rooms.value = []

  if (!buildingId) return

  try {
    const [floorsRes, roomsRes] = await Promise.all([
      sqBuildingApi.listFloors(buildingId),
      sqBuildingApi.listRooms(buildingId)
    ])
    if (floorsRes.data?.code === 0) {
      floors.value = floorsRes.data.data.items || []
    }
    if (roomsRes.data?.code === 0) {
      rooms.value = roomsRes.data.data.items || []
    }
  } catch {
    ElMessage.error('加载楼层/寝室数据失败')
  }
}

// 扣分项操作
function addDeduction() {
  form.deductions.push({ item: '', deduction: 0 })
}

function removeDeduction(index) {
  form.deductions.splice(index, 1)
}

// 提交
async function handleSubmit() {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitting.value = true
  try {
    const data = {
      inspection_type: form.inspection_type,
      building_id: form.building_id,
      floor_id: form.floor_id || undefined,
      room_id: form.room_id || undefined,
      inspected_at: form.inspected_at,
      summary: form.summary,
      deductions: form.deductions.filter(d => d.item.trim() !== ''),
      score: computedScore.value
    }
    await sqInspectionApi.create(data)
    ElMessage.success('提交成功')
    router.push('/sq/inspection')
  } catch {
    ElMessage.error('提交失败')
  } finally {
    submitting.value = false
  }
}

// 初始化
onMounted(async () => {
  try {
    const data = await sqBuildingApi.list()
    buildings.value = data?.items || []
  } catch {
    // 静默处理
  }
  // 默认一行扣分项
  addDeduction()
})
</script>

<style scoped>
.score-preview {
  font-size: var(--sh-text-2xl);
  font-weight: 700;
}
.score-green {
  color: var(--sh-success);
}
.score-orange {
  color: var(--sh-warning);
}
.score-red {
  color: var(--sh-danger);
}
</style>
