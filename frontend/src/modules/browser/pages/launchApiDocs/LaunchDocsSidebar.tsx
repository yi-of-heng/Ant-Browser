import { useMemo, useState } from 'react'
import { FileText, Search, Sparkles, X } from 'lucide-react'
import type { LaunchDocGroup } from './catalog'

interface LaunchDocsSidebarProps {
  groups: LaunchDocGroup[]
  activeId: string
  onSelect: (id: string) => void
}

export function LaunchDocsSidebar({
  groups,
  activeId,
  onSelect,
}: LaunchDocsSidebarProps) {
  const [query, setQuery] = useState('')
  const normalizedQuery = query.trim().toLowerCase()
  const visibleGroups = useMemo(() => groups
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => !normalizedQuery
        || `${item.label} ${item.summary} ${group.label}`.toLowerCase().includes(normalizedQuery)),
    }))
    .filter((group) => group.items.length > 0), [groups, normalizedQuery])
  const totalDocs = groups.reduce((total, group) => total + group.items.length, 0)

  return (
    <div className="space-y-5">
      <div className="flex items-start gap-3 px-1">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-[var(--color-accent)] text-white shadow-[var(--shadow-sm)]">
          <Sparkles className="h-4 w-4" />
        </div>
        <div className="min-w-0">
          <p className="text-sm font-semibold text-[var(--color-text-primary)]">开发者中心</p>
          <p className="mt-0.5 text-xs leading-5 text-[var(--color-text-muted)]">API、CLI、MCP 与自动化指南</p>
        </div>
        <span className="ml-auto rounded-full border border-[var(--color-border-default)] bg-[var(--color-bg-subtle)] px-2 py-1 text-[10px] font-semibold text-[var(--color-text-muted)]">
          1.8
        </span>
      </div>

      <label className="relative block">
        <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--color-text-muted)]" />
        <input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder={`搜索文档（${totalDocs} 篇）`}
          className="h-10 w-full rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-subtle)] pl-9 pr-9 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-accent)] focus:ring-2 focus:ring-[var(--color-accent-muted)]"
        />
        {query && (
          <button
            type="button"
            aria-label="清除搜索"
            onClick={() => setQuery('')}
            className="absolute right-2 top-1/2 flex h-6 w-6 -translate-y-1/2 items-center justify-center rounded-md text-[var(--color-text-muted)] hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-text-primary)]"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        )}
      </label>

      <nav className="space-y-5">
        {visibleGroups.map((group) => (
          <section key={group.id} className="space-y-1">
            <div className="flex items-center justify-between px-2 pb-1">
              <p className="text-[11px] font-semibold uppercase tracking-[0.16em] text-[var(--color-text-muted)]">
                {group.label}
              </p>
              <span className="text-[10px] tabular-nums text-[var(--color-text-muted)]">{group.items.length}</span>
            </div>
            {group.items.map((item) => {
              const isActive = activeId === item.id
              return (
                <button
                  key={item.id}
                  onClick={() => onSelect(item.id)}
                  className={[
                    'group relative w-full rounded-xl border px-3 py-2.5 text-left transition-all',
                    isActive
                      ? 'border-[var(--color-accent)] bg-[var(--color-accent-muted)] shadow-[var(--shadow-sm)]'
                      : 'border-transparent hover:border-[var(--color-border-muted)] hover:bg-[var(--color-bg-muted)]',
                  ].join(' ')}
                >
                  {isActive && <span className="absolute bottom-2 left-0 top-2 w-0.5 rounded-full bg-[var(--color-accent)]" />}
                  <div className="flex items-start gap-2">
                    <FileText className={`mt-0.5 h-3.5 w-3.5 shrink-0 ${isActive ? 'text-[var(--color-accent)]' : 'text-[var(--color-text-muted)] group-hover:text-[var(--color-text-secondary)]'}`} />
                    <div className={`min-w-0 text-sm font-medium ${isActive ? 'text-[var(--color-text-primary)]' : 'text-[var(--color-text-secondary)]'}`}>
                      <div>{item.label}</div>
                      <div className="mt-0.5 line-clamp-1 text-[11px] font-normal leading-4 text-[var(--color-text-muted)]">
                        {item.summary}
                      </div>
                    </div>
                  </div>
                </button>
              )
            })}
          </section>
        ))}
      </nav>

      {visibleGroups.length === 0 && (
        <div className="rounded-xl border border-dashed border-[var(--color-border-default)] px-3 py-5 text-center text-xs text-[var(--color-text-muted)]">
          没有匹配的文档
        </div>
      )}
    </div>
  )
}
