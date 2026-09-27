export type GardenStatus = 'active' | 'removed'

export const GardenStatusMap: Record<GardenStatus, string> = {
  active: '在养',
  removed: '已移除',
}

export const GARDEN_STATUSES = Object.keys(GardenStatusMap) as GardenStatus[]
