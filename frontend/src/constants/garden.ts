export type GardenLocation = 'indoor' | 'balcony'

export const GardenLocationMap: Record<GardenLocation, string> = {
  indoor: '室内',
  balcony: '阳台',
}

export const GARDEN_LOCATIONS = Object.keys(GardenLocationMap) as GardenLocation[]

export function gardenLocationText(loc: string): string {
  return GardenLocationMap[loc as GardenLocation] || loc || '-'
}
