<template>
  <div class="page">
    <h1>我的花园</h1>

    <el-row :gutter="16" class="stats">
      <el-col :xs="8">
        <el-card shadow="hover">
          <div class="stat-num">{{ gardenItems.length }}</div>
          <div class="stat-label">当前清单植株</div>
        </el-card>
      </el-col>
      <el-col :xs="8">
        <el-card shadow="hover">
          <div class="stat-num">{{ todoCount }}</div>
          <div class="stat-label">待办提醒</div>
        </el-card>
      </el-col>
      <el-col :xs="8">
        <el-card shadow="hover">
          <div class="stat-num">{{ totalRemaining }}</div>
          <div class="stat-label">位置剩余空位</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="block">
      <template #header>
        <div class="card-header">
          <span>花园清单</span>
          <div class="header-actions">
            <el-radio-group v-model="locationFilter" @change="reload">
              <el-radio-button :value="0">全部</el-radio-button>
              <el-radio-button v-for="loc in locations" :key="loc.id" :value="loc.id">
                {{ loc.name }}（余 {{ loc.remaining }}）
              </el-radio-button>
            </el-radio-group>
            <el-button size="small" @click="openCapacityDialog">容量设置</el-button>
          </div>
        </div>
      </template>

      <div v-if="selected.length" class="move-bar">
        <span>已选 {{ selected.length }} 株</span>
        <el-select v-model="batchTarget" placeholder="搬到位置" style="width: 180px">
          <el-option
            v-for="loc in locations"
            :key="loc.id"
            :label="`${loc.name}（余 ${loc.remaining}）`"
            :value="loc.id"
            :disabled="loc.remaining <= 0"
          />
        </el-select>
        <el-button type="primary" size="small" :disabled="!batchTarget" @click="batchMove">批量搬位</el-button>
      </div>

      <el-table :data="gardenItems" empty-text="花园还是空的，去品种库添加吧" @selection-change="selected = $event">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="garden_code" label="园内编号" width="110" />
        <el-table-column label="植物">
          <template #default="{ row }">
            <span class="garden-name">{{ row.nickname || row.plant_species_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="位置" width="100">
          <template #default="{ row }">{{ row.location_name || '未分配' }}</template>
        </el-table-column>
        <el-table-column label="拥有时间" width="120">
          <template #default="{ row }">{{ formatDate(row.owned_since) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button size="small" @click="openMove([row.id])">搬位</el-button>
            <el-button size="small" type="danger" @click="remove(row.id)">移除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-row :gutter="16" class="block">
      <el-col :xs="24" :md="16">
        <el-card>
          <template #header>我的养护提醒</template>
          <el-form inline>
            <el-form-item label="植株">
              <el-select v-model="reminderForm.garden_id" placeholder="通用任务可不选" clearable style="width: 200px">
                <el-option
                  v-for="g in gardenItems"
                  :key="g.id"
                  :label="`${g.nickname || g.garden_code}（${g.garden_code}）`"
                  :value="g.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="任务">
              <el-input v-model="reminderForm.task_title" placeholder="如：给月季施肥" style="width: 180px" />
            </el-form-item>
            <el-form-item label="日期">
              <el-date-picker v-model="reminderForm.remind_date" type="date" value-format="YYYY-MM-DD" style="width: 150px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingReminder" @click="createPlantReminder">保存</el-button>
            </el-form-item>
          </el-form>
          <ReminderList :reminders="reminders" @done="markDone" @remove="removeReminder" />
        </el-card>

        <el-card class="block">
          <template #header>历史记录</template>
          <el-tabs v-model="historyTab">
            <el-tab-pane label="搬位记录" name="moves">
              <el-table :data="moveRecords" empty-text="暂无搬位记录">
                <el-table-column prop="garden_code" label="园内编号" width="110" />
                <el-table-column prop="plant_nickname" label="植株" />
                <el-table-column label="搬位" width="160">
                  <template #default="{ row }">{{ row.from_location }} → {{ row.to_location }}</template>
                </el-table-column>
                <el-table-column label="时间" width="150">
                  <template #default="{ row }">{{ formatDateTime(row.moved_at) }}</template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
            <el-tab-pane label="已移除植株" name="removed">
              <el-table :data="removedItems" empty-text="暂无已移除植株">
                <el-table-column prop="garden_code" label="园内编号" width="110" />
                <el-table-column prop="nickname" label="昵称" />
                <el-table-column label="状态" width="90">
                  <template #default="{ row }">{{ GardenStatusMap[row.status as GardenStatus] }}</template>
                </el-table-column>
                <el-table-column label="移除时间" width="150">
                  <template #default="{ row }">{{ formatDateTime(row.removed_at) }}</template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
          </el-tabs>
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

    <el-dialog v-model="moveDialog" title="植株搬位" width="420px">
      <p>将 {{ movingIds.length }} 株植株搬到：</p>
      <el-radio-group v-model="moveTarget" class="move-options">
        <el-radio v-for="loc in locations" :key="loc.id" :value="loc.id" :disabled="loc.remaining <= 0">
          {{ loc.name }}（剩余 {{ loc.remaining }} 个空位）
        </el-radio>
      </el-radio-group>
      <template #footer>
        <el-button @click="moveDialog = false">取消</el-button>
        <el-button type="primary" :disabled="!moveTarget" :loading="moving" @click="confirmMove">确定搬位</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="capacityDialog" title="位置容量设置" width="420px">
      <el-form label-width="60px">
        <el-form-item v-for="item in capacityForm" :key="item.id" :label="item.name">
          <el-input-number v-model="item.capacity" :min="item.used" />
          <span class="used-hint">当前已用 {{ item.used }}</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="capacityDialog = false">取消</el-button>
        <el-button type="primary" :loading="savingCapacity" @click="saveCapacities">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import ReminderList from '@/components/common/ReminderList.vue'
import {
  listGardens,
  listRemovedGardens,
  listMoveRecords,
  listLocations,
  moveGardens,
  removeGarden,
  updateLocationCapacity,
} from '@/api/garden'
import { listFavorites } from '@/api/favorite'
import { listReminders, createReminder, deleteReminder, updateReminderStatus } from '@/api/reminder'
import { FavoriteTargetTypeMap, type Favorite, type FavoriteTargetType } from '@/constants/favorite'
import { GardenStatusMap, type GardenStatus } from '@/constants/garden'
import { formatDate, formatDateTime } from '@/utils/dateFormat'
import type { CareReminder, GardenLocation, GardenMoveRecord, UserGarden } from '@/types/api'

const gardenItems = ref<UserGarden[]>([])
const removedItems = ref<UserGarden[]>([])
const moveRecords = ref<GardenMoveRecord[]>([])
const locations = ref<GardenLocation[]>([])
const favorites = ref<Favorite[]>([])
const reminders = ref<CareReminder[]>([])

const locationFilter = ref(0)
const selected = ref<UserGarden[]>([])
const batchTarget = ref<number>()
const historyTab = ref('moves')

const moveDialog = ref(false)
const moveTarget = ref<number>()
const movingIds = ref<number[]>([])
const moving = ref(false)

const capacityDialog = ref(false)
const capacityForm = ref<{ id: number; name: string; capacity: number; used: number }[]>([])
const savingCapacity = ref(false)

const savingReminder = ref(false)
const reminderForm = reactive({ garden_id: undefined as number | undefined, task_title: '', remind_date: '' })

const todoCount = computed(() => reminders.value.filter((r) => r.status !== 'done').length)
const totalRemaining = computed(() => {
  if (locationFilter.value) {
    const loc = locations.value.find((l) => l.id === locationFilter.value)
    return loc ? loc.remaining : 0
  }
  return locations.value.reduce((sum, loc) => sum + loc.remaining, 0)
})

onMounted(async () => {
  await Promise.all([reload(), reloadHistory(), reloadLocations()])
  favorites.value = await listFavorites()
})

// 按位置筛选时清单、待办和数量一起变化
async function reload() {
  const locationId = locationFilter.value || undefined
  const [items, rems] = await Promise.all([listGardens(locationId), listReminders(undefined, locationId)])
  gardenItems.value = items
  reminders.value = rems
}

async function reloadLocations() {
  locations.value = await listLocations()
}

async function reloadHistory() {
  const [removed, moves] = await Promise.all([listRemovedGardens(), listMoveRecords()])
  removedItems.value = removed
  moveRecords.value = moves
}

function openMove(ids: number[]) {
  movingIds.value = ids
  moveTarget.value = undefined
  moveDialog.value = true
}

async function confirmMove() {
  if (!moveTarget.value) return
  moving.value = true
  try {
    await moveGardens(movingIds.value.map((id) => ({ garden_id: id, to_location_id: moveTarget.value! })))
    ElMessage.success('搬位成功')
    moveDialog.value = false
    await Promise.all([reload(), reloadLocations(), reloadHistory()])
  } finally {
    moving.value = false
  }
}

async function batchMove() {
  if (!batchTarget.value || !selected.value.length) return
  await moveGardens(selected.value.map((g) => ({ garden_id: g.id, to_location_id: batchTarget.value! })))
  ElMessage.success('批量搬位成功')
  batchTarget.value = undefined
  await Promise.all([reload(), reloadLocations(), reloadHistory()])
}

async function remove(id: number) {
  await removeGarden(id)
  ElMessage.success('已从我的花园移除')
  await Promise.all([reload(), reloadLocations(), reloadHistory()])
}

async function createPlantReminder() {
  if (!reminderForm.task_title || !reminderForm.remind_date) {
    ElMessage.warning('请填写任务与日期')
    return
  }
  savingReminder.value = true
  try {
    await createReminder({
      garden_id: reminderForm.garden_id,
      task_title: reminderForm.task_title,
      remind_date: reminderForm.remind_date,
    })
    ElMessage.success('养护提醒已创建')
    reminderForm.task_title = ''
    reminderForm.remind_date = ''
    await reload()
  } finally {
    savingReminder.value = false
  }
}

async function markDone(id: number) {
  await updateReminderStatus(id, 'done')
  await reload()
}

async function removeReminder(id: number) {
  await deleteReminder(id)
  await reload()
}

function openCapacityDialog() {
  capacityForm.value = locations.value.map((loc) => ({ id: loc.id, name: loc.name, capacity: loc.capacity, used: loc.used }))
  capacityDialog.value = true
}

async function saveCapacities() {
  savingCapacity.value = true
  try {
    for (const item of capacityForm.value) {
      await updateLocationCapacity(item.id, item.capacity)
    }
    ElMessage.success('位置容量已更新')
    capacityDialog.value = false
    await Promise.all([reload(), reloadLocations()])
  } finally {
    savingCapacity.value = false
  }
}
</script>

<style scoped>
.page { max-width: 1200px; margin: 0 auto; }
.block { margin-top: 16px; }
.garden-name { font-weight: 600; }
.stats { margin-bottom: 4px; }
.stat-num { font-size: 28px; font-weight: 700; color: #2c6e49; }
.stat-label { color: #888; margin-top: 4px; }
.card-header { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; }
.header-actions { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.move-bar { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.move-options { display: flex; flex-direction: column; gap: 8px; }
.used-hint { margin-left: 12px; color: #999; font-size: 12px; }
</style>
