import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listReminders, updateReminderStatus, createReminder } from '@/api/reminder'
import type { CareReminder } from '@/types/api'

export const useReminderStore = defineStore('reminder', () => {
  const reminders = ref<CareReminder[]>([])

  async function load(status?: string, location?: string) {
    reminders.value = await listReminders(status, location)
  }

  async function create(payload: { plant_species_id?: number; garden_id?: number; task_title: string; remind_date: string; frequency?: string }) {
    await createReminder(payload)
    await load()
  }

  async function setStatus(id: number, status: string) {
    await updateReminderStatus(id, status)
    await load()
  }

  return { reminders, load, create, setStatus }
})
