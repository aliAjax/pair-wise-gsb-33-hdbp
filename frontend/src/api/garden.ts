import request from '@/utils/request'
import type { GardenLocationStat, GardenMove, UserGarden } from '@/types/api'

export function listGardens(location?: string) {
  return request.get<never, UserGarden[]>('/gardens', { params: { location: location || undefined } })
}

export function addGarden(payload: { plant_species_id: number; nickname?: string; owned_since?: string; location?: string }) {
  return request.post<never, UserGarden>('/gardens', payload)
}

export function bindReminder(id: number, careReminderId: number) {
  return request.put<never, UserGarden>(`/gardens/${id}/reminder`, { care_reminder_id: careReminderId })
}

export function removeGarden(id: number) {
  return request.delete<never, { removed: boolean }>(`/gardens/${id}`)
}

export function listGardenLocations() {
  return request.get<never, GardenLocationStat[]>('/gardens/locations')
}

export function updateLocationCapacity(location: string, capacity: number) {
  return request.put<never, GardenLocationStat>(`/gardens/locations/${location}`, { capacity })
}

export function moveGarden(id: number, location: string) {
  return request.put<never, UserGarden>(`/gardens/${id}/location`, { location })
}

export function batchMoveGarden(moves: { garden_id: number; location: string }[]) {
  return request.post<never, { moved: number }>('/gardens/moves', { moves })
}

export function listGardenMoves(gardenId?: number) {
  return request.get<never, GardenMove[]>('/gardens/moves', { params: { garden_id: gardenId || undefined } })
}
