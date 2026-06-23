<template>
  <div class="incident-report">
    <el-card shadow="never">
      <template #header>
        <span>异常事件上报</span>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px" style="max-width: 640px">
        <el-form-item label="事件等级" prop="incident_level">
          <el-radio-group v-model="form.incident_level">
            <el-radio label="L1">L1</el-radio>
            <el-radio label="L2">L2</el-radio>
            <el-radio label="L3">L3</el-radio>
            <el-radio label="L4">L4</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-alert
          v-if="form.incident_level === 'L4'"
          title="L4 级事件将触发应急通知，请谨慎上报"
          type="error"
          show-icon
          :closable="false"
          style="margin-bottom: 18px; max-width: 520px"
        />

        <el-form-item label="事件类型" prop="incident_type">
          <el-input v-model="form.incident_type" placeholder="请输入事件类型" />
        </el-form-item>

        <el-form-item label="发生时间" prop="occurred_at">
          <el-date-picker
            v-model="form.occurred_at"
            type="datetime"
            placeholder="选择发生时间"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="楼栋" prop="building_id">
          <el-select v-model="form.building_id" placeholder="请选择楼栋" style="width: 100%" @change="onBuildingChange">
            <el-option
              v-for="b in buildings"
              :key="b.id"
              :label="b.name"
              :value="b.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="楼层">
          <el-select v-model="form.floor_id" placeholder="请选择楼层" clearable style="width: 100%">
            <el-option
              v-for="f in floors"
              :key="f.id"
              :label="f.name"
              :value="f.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="寝室">
          <el-select v-model="form.room_id" placeholder="请选择寝室" clearable style="width: 100%">
            <el-option
              v-for="r in rooms"
              :key="r.id"
              :label="r.name"
              :value="r.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="地点详情">
          <el-input v-model="form.location_detail" placeholder="请输入地点详情" />
        </el-form-item>

        <el-form-item label="涉事学生">
          <el-input v-model="involvedStudentInput" placeholder="输入学生ID，逗号分隔" />
        </el-form-item>

        <el-form-item label="证人">
          <el-input v-model="witnessUserInput" placeholder="输入用户ID，逗号分隔" />
        </el-form-item>

        <el-form-item label="初步处置" prop="initial_action">
          <el-input v-model="form.initial_action" type="textarea" :rows="4" placeholder="请输入初步处置说明" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">提交</el-button>
          <el-button @click="router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { sqIncidentApi, sqBuildingApi } from '@/api/sq'

const router = useRouter()

const formRef = ref(null)
const submitting = ref(false)

const form = reactive({
  incident_level: '',
  incident_type: '',
  occurred_at: '',
  building_id: '',
  floor_id: '',
  room_id: '',
  location_detail: '',
  involved_student_ids: [],
  witness_user_ids: [],
  initial_action: ''
})

const involvedStudentInput = ref('')
const witnessUserInput = ref('')

const rules = {
  incident_level: [{ required: true, message: '请选择事件等级', trigger: 'change' }],
  incident_type: [{ required: true, message: '请输入事件类型', trigger: 'blur' }],
  occurred_at: [{ required: true, message: '请选择发生时间', trigger: 'change' }],
  building_id: [{ required: true, message: '请选择楼栋', trigger: 'change' }],
  initial_action: [{ required: true, message: '请输入初步处置说明', trigger: 'blur' }]
}

// 楼栋/楼层/寝室
const buildings = ref([])
const floors = ref([])
const rooms = ref([])

const fetchBuildings = async () => {
  try {
    const data = await sqBuildingApi.list()
    buildings.value = data?.items || data || []
  } catch (e) {
    console.error('获取楼栋列表失败', e)
  }
}

const onBuildingChange = async (buildingId) => {
  form.floor_id = ''
  form.room_id = ''
  floors.value = []
  rooms.value = []
  if (!buildingId) return
  try {
    const [floorsData, roomsData] = await Promise.all([
      sqBuildingApi.listFloors(buildingId),
      sqBuildingApi.listRooms(buildingId)
    ])
    floors.value = floorsData?.items || floorsData || []
    rooms.value = roomsData?.items || roomsData || []
  } catch (e) {
    console.error('获取楼层/寝室列表失败', e)
  }
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  // 解析逗号分隔的ID
  form.involved_student_ids = involvedStudentInput.value
    ? involvedStudentInput.value.split(',').map(s => s.trim()).filter(Boolean)
    : []
  form.witness_user_ids = witnessUserInput.value
    ? witnessUserInput.value.split(',').map(s => s.trim()).filter(Boolean)
    : []

  submitting.value = true
  try {
    await sqIncidentApi.create(form)
    ElMessage.success('上报成功')
    router.push('/sq/incident')
  } catch (e) {
    ElMessage.error('上报失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchBuildings()
})
</script>

<style scoped>
.incident-report {
  padding: var(--sh-space-lg);
}
</style>
