import type { BrowserProfile } from '../types'

/**
 * Groups active browser profiles by their persisted proxy-pool node ID.
 *
 * A custom per-profile proxy has no pool ID and is intentionally omitted: it
 * is not a connection to a reusable proxy-pool node.
 */
export function buildProxyUsageByID(profiles: BrowserProfile[]): Record<string, BrowserProfile[]> {
  return profiles.reduce<Record<string, BrowserProfile[]>>((usage, profile) => {
    const proxyId = (profile.proxyId || '').trim()
    if (!proxyId || profile.deletedAt) return usage
    const linkedProfiles = usage[proxyId] || []
    linkedProfiles.push(profile)
    usage[proxyId] = linkedProfiles
    return usage
  }, {})
}

export function proxyUsageSummary(profiles: BrowserProfile[], maxNames = 2): string {
  if (profiles.length === 0) return '未绑定实例'
  const names = profiles
    .map(profile => profile.profileName || profile.profileId)
    .filter(Boolean)
  const shown = names.slice(0, maxNames)
  const remaining = names.length - shown.length
  return remaining > 0 ? `${shown.join('、')} 等 ${names.length} 个实例` : shown.join('、')
}
