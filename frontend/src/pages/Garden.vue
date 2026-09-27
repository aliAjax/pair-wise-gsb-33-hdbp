<template>
  <div class="page">
    <h1>我的花园</h1>

    <el-row :gutter="16" class="stat-row">
      <el-col v-for="s in locationStats" :key="s.location" :xs="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-head">
            <span class="stat-title">{{ gardenLocationText(s.location) }}</span>
            <el-button size="small" text type="primary" @click="openCapacity(s)">设置容量</el-button>
          </div>
          <div class="stat-nums">{{ s.used }}<span class="stat-cap"> / {{ s.capacity }}</span></div>
          <div class="stat-remaining" :class="{ full: s.remaining <= 0 }">余量 {{ s.remaining }}</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="filter-card">
          <span class="filter-label">按位置筛选</span>
          <el-radio-group v-model="locationFilter" @change="reload">
            <el-radio-button value="">全部</el-radio-button>
            <el-radio-button v-for="loc in GARDEN_LOCATIONS" :key="loc" :value="loc">
              {{ gardenLocationText(loc) }}
            </el-radio-button>
          </el-radio-group>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :xs="24" :md="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>植株档案</span>
              <el-button size="small" type="primary" :disabled="!selected.length" @click="openBatchMove">
                批量搬位（{{ selected.length }}）
              </el-button>
            </div>
          </template>
          <el-table :data="gardenItems" empty-text="花园还是空的，去品种库添加吧" @selection-change="onSelectionChange">
            <el-table-column type="selection" width="45" />
            <el-table-column prop="garden_no" label="编号" width="90" />
            <el-table-column label="昵称" min-width="120">
              <template #default="{ row }">
                <span class="garden-name">{{ row.nickname || row.garden_no }}</span>
              </template>
            </el-table-column>
            <el-table-column label="位置" width="90">
              <template #default="{ row }">
                <el-tag :type="row.location === 'balcony' ? 'warning' : 'success'">
                  {{ gardenLocationText(row.location) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="拥有时间" width="110">
              <template #default="{ row }">{{ formatDate(row.owned_since) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="230">
              <template #default="{ row }">
                <el-button size="small" @click="openMove(row)">搬位</el-button>
                <el-button size="small" type="primary" @click="openReminder(row)">提醒</el-button>
                <el-button size="small" type="danger" @click="remove(row)">移除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="block">
          <template #header>
            <div class="card-header">
              <span>养护待办</span>
              <el-button size="small" type="primary" @click="openReminder()">新建提醒</el-button>
            </div>
          </template>
          <ReminderList :reminders="reminders" @done="markDone" @remove="removeReminder" />
        </el-card>

        <el-card class="block">
          <template #header>搬位记录</template>
          <el-table :data="moves" empty-text="暂无搬位记录">
            <el-table-column label="时间" width="150">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column prop="garden_no" label="编号" width="90" />
            <el-table-column prop="nickname" label="昵称" min-width="110" />
            <el-table-column label="搬位" min-width="140">
              <template #default="{ row }">
                {{ gardenLocationText(row.from_location) }} → {{ gardenLocationText(row.to_location) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :xs="24" :md="8">
        <el-card>
          <template #header>我的收藏</template>
          <el-table :data="favorites" empty-text="暂无收藏">
            <el-table-column prop="target_type" label="类型" width="80">
              <template #default="{ row }">{{ FavoriteTargetTypeMap[row.target_type as FavoriteTargetType] }}</template>
            </el-table-column>
            <el-table-column prop="target_id" label="目标 ID" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="moveDialog" title="搬位" width="420px">
      <p class="dialog-tip">
        将「{{ moveTarget?.nickname || moveTarget?.garden_no }}」（{{ moveTarget?.garden_no }}）搬到：
      </p>
      <el-radio-group v-model="moveLocation">
        <el-radio v-for="loc in GARDEN_LOCATIONS" :key="loc" :value="loc" :disabled="loc === moveTarget?.location">
          {{ gardenLocationText(loc) }}（余量 {{ remainingOf(loc) }}）
        </el-radio>
      </el-radio-group>
      <template #footer>
        <el-button @click="moveDialog = false">取消</el-button>
        <el-button type="primary" :disabled="!moveLocation" @click="confirmMove">确认搬位</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchDialog" title="批量搬位" width="420px">
      <p class="dialog-tip">将选中的 {{ selected.length }} 株植物搬到：</p>
      <el-radio-group v-model="batchLocation">
        <el-radio v-for="loc in GARDEN_LOCATIONS" :key="loc" :value="loc">
          {{ gardenLocationText(loc) }}（余量 {{ remainingOf(loc) }}）
        </el-radio>
      </el-radio-group>
      <template #footer>
        <el-button @click="batchDialog = false">取消</el-button>
        <el-button type="primary" :disabled="!batchLocation" @click="confirmBatchMove">确认搬位</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="reminderDialog" title="新建养护提醒" width="480px">
      <el-form label-width="80px">
        <el-form-item label="植株">
          <el-select v-model="reminderForm.garden_id" placeholder="选择植株" style="width: 100%">
            <el-option
              v-for="g in gardenItems"
              :key="g.id"
              :value="g.id"
              :label="`${g.nickname || g.garden_no}（${g.garden_no} · ${gardenLocationText(g.location)}）`"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="任务">
          <el-input v-model="reminderForm.task_title" placeholder="如：给月季施肥" />
        </el-form-item>
        <el-form-item label="日期">
          <el-date-picker v-model="reminderForm.remind_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="频率">
          <el-select v-model="reminderForm.frequency" clearable placeholder="一次性" style="width: 100%">
            <el-option value="daily" label="每日" />
            <el-option value="weekly" label="每周" />
            <el-option value="monthly" label="每月" />
            <el-option value="yearly" label="每年" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reminderDialog = false">取消</el-button>
        <el-button type="primary" @click="saveReminder">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="capacityDialog" title="设置容量" width="360px">
      <p class="dialog-tip">{{ gardenLocationText(capacityLocation) }}可容纳的植株数量：</p>
      <el-input-number v-model="capacityValue" :min="1" :max="999" />
      <template #footer>
        <el-button @click="capacityDialog = false">取消</el-button>
        <el-button type="primary" @click="saveCapacity">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ReminderList from '@/components/common/ReminderList.vue'
import {
  batchMoveGarden,
  listGardenLocations,
  listGardenMoves,
  listGardens,
  moveGarden,
  removeGarden,
  updateLocationCapacity,
} from '@/api/garden'
import { listFavorites } from '@/api/favorite'
import { createReminder, deleteReminder, listReminders, updateReminderStatus } from '@/api/reminder'
import { FavoriteTargetTypeMap, type Favorite, type FavoriteTargetType } from '@/constants/favorite'
import { GARDEN_LOCATIONS, gardenLocationText } from '@/constants/garden'
import { formatDate, formatDateTime } from '@/utils/dateFormat'
import type { CareReminder, GardenLocationStat, GardenMove, UserGarden } from '@/types/api'

const gardenItems = ref<UserGarden[]>([])
const favorites = ref<Favorite[]>([])
const reminders = ref<CareReminder[]>([])
const locationStats = ref<GardenLocationStat[]>([])
const moves = ref<GardenMove[]>([])
const locationFilter = ref('')
const selected = ref<UserGarden[]>([])

// 清单、待办、数量（容量统计）与搬位记录一起刷新，保证筛选后数据一致
async function reload() {
  const [g, r, stats, mv] = await Promise.all([
    listGardens(locationFilter.value || undefined),
    listReminders(undefined, locationFilter.value || undefined),
    listGardenLocations(),
    listGardenMoves(),
  ])
  gardenItems.value = g
  reminders.value = r
  locationStats.value = stats
  moves.value = mv
}

onMounted(async () => {
  await reload()
  favorites.value = await listFavorites()
})

function onSelectionChange(rows: UserGarden[]) {
  selected.value = rows
}

function remainingOf(loc: string): number {
  return locationStats.value.find((s) => s.location === loc)?.remaining ?? 0
}

const moveDialog = ref(false)
const moveTarget = ref<UserGarden | null>(null)
const moveLocation = ref('')
function openMove(row: UserGarden) {
  moveTarget.value = row
  moveLocation.value = ''
  moveDialog.value = true
}
async function confirmMove() {
  if (!moveTarget.value) return
  await moveGarden(moveTarget.value.id, moveLocation.value)
  moveDialog.value = false
  ElMessage.success('已搬位')
  await reload()
}

const batchDialog = ref(false)
const batchLocation = ref('')
function openBatchMove() {
  batchLocation.value = ''
  batchDialog.value = true
}
async function confirmBatchMove() {
  // 任一位置放不下时后端会整批拒绝并在消息里说明缺口，拦截器负责弹出
  await batchMoveGarden(selected.value.map((g) => ({ garden_id: g.id, location: batchLocation.value })))
  batchDialog.value = false
  ElMessage.success('批量搬位完成')
  await reload()
}

const reminderDialog = ref(false)
const reminderForm = reactive<{ garden_id?: number; task_title: string; remind_date: string; frequency: string }>({
  garden_id: undefined,
  task_title: '',
  remind_date: '',
  frequency: '',
})
function openReminder(row?: UserGarden) {
  reminderForm.garden_id = row?.id
  reminderForm.task_title = ''
  reminderForm.remind_date = ''
  reminderForm.frequency = ''
  reminderDialog.value = true
}
async function saveReminder() {
  if (!reminderForm.garden_id || !reminderForm.task_title || !reminderForm.remind_date) {
    ElMessage.warning('请选择植株并填写任务与日期')
    return
  }
  await createReminder({
    garden_id: reminderForm.garden_id,
    task_title: reminderForm.task_title,
    remind_date: reminderForm.remind_date,
    frequency: reminderForm.frequency || undefined,
  })
  reminderDialog.value = false
  ElMessage.success('养护提醒已创建')
  await reload()
}

const capacityDialog = ref(false)
const capacityLocation = ref('')
const capacityValue = ref(1)
function openCapacity(s: GardenLocationStat) {
  capacityLocation.value = s.location
  capacityValue.value = s.capacity
  capacityDialog.value = true
}
async function saveCapacity() {
  await updateLocationCapacity(capacityLocation.value, capacityValue.value)
  capacityDialog.value = false
  ElMessage.success('容量已更新')
  await reload()
}

async function remove(row: UserGarden) {
  await ElMessageBox.confirm(`移除「${row.nickname || row.garden_no}」？搬位记录与已完成提醒仍可查询。`, '移除植株', {
    type: 'warning',
    confirmButtonText: '移除',
    cancelButtonText: '取消',
  })
  await removeGarden(row.id)
  ElMessage.success('已移除')
  await reload()
}

async function markDone(id: number) {
  await updateReminderStatus(id, 'done')
  await reload()
}
async function removeReminder(id: number) {
  await deleteReminder(id)
  ElMessage.success('已删除提醒')
  await reload()
}
</script>

<style scoped>
.page { max-width: 1200px; margin: 0 auto; }
.block { margin-top: 16px; }
.garden-name { font-weight: 600; }
.stat-row { margin-bottom: 16px; }
.stat-card :deep(.el-card__body) { padding: 14px 16px; }
.stat-head { display: flex; justify-content: space-between; align-items: center; }
.stat-title { font-weight: 600; }
.stat-nums { font-size: 26px; font-weight: 700; margin-top: 4px; }
.stat-cap { font-size: 14px; color: #999; font-weight: 400; }
.stat-remaining { color: #67c23a; font-size: 13px; }
.stat-remaining.full { color: #f56c6c; }
.filter-card :deep(.el-card__body) { display: flex; align-items: center; gap: 12px; }
.filter-label { color: #666; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.dialog-tip { margin-top: 0; }
</style>
