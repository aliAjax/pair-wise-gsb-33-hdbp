import request from '@/utils/request'
import type { GardenLocation, GardenMoveRecord, UserGarden } from '@/types/api'

export function listGardens(locationId?: number) {
  return request.get<never, UserGarden[]>('/gardens', { params: { location_id: locationId || undefined } })
}

export function listRemovedGardens() {
  return request.get<never, UserGarden[]>('/gardens/removed')
}

export function listMoveRecords() {
  return request.get<never, GardenMoveRecord[]>('/gardens/moves')
}

export function addGarden(payload: { plant_species_id: number; nickname?: string; owned_since?: string; location_id?: number }) {
  return request.post<never, UserGarden>('/gardens', payload)
}

export function moveGardens(moves: { garden_id: number; to_location_id: number }[]) {
  return request.post<never, UserGarden[]>('/gardens/move', { moves })
}

export function listLocations() {
  return request.get<never, GardenLocation[]>('/garden-locations')
}

export function updateLocationCapacity(id: number, capacity: number) {
  return request.put<never, GardenLocation>(`/garden-locations/${id}`, { capacity })
}

export function bindReminder(id: number, careReminderId: number) {
  return request.put<never, UserGarden>(`/gardens/${id}/reminder`, { care_reminder_id: careReminderId })
}

export function removeGarden(id: number) {
  return request.delete<never, { removed: boolean }>(`/gardens/${id}`)
}
