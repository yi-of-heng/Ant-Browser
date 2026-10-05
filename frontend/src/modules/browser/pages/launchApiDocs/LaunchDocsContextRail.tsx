import { CheckCircle2, ChevronRight, Copy, ExternalLink, KeyRound, RefreshCw, Server, Terminal } from 'lucide-react'
import { Button, toast } from '../../../../shared/components'
import type { LaunchServerInfo } from '../../api'
import type { LaunchDocItem } from './catalog'

interface LaunchDocsContextRailProps {
  currentGroupLabel: string
  currentDocId: string
  currentDocLabel: string
  launchBaseUrl: string
  launchServerReady: boolean
  launchContextLoading: boolean
  apiAuth: LaunchServerInfo['apiAuth']
  relatedDocs: LaunchDocItem[]
  onRefresh: () => void
  onSelectDoc: (id: string) => void
}

export function LaunchDocsContextRail({
  currentGroupLabel,
  currentDocId,
  currentDocLabel,
  launchBaseUrl,
  launchServerReady,
  launchContextLoading,
  apiAuth,
  relatedDocs,
  onRefresh,
  onSelectDoc,
}: LaunchDocsContextRailProps) {
  const copyBaseUrl = async () => {
    try {
      await navigator.clipboard.writeText(launchBaseUrl)
      toast.success('Launch API 地址已复制')
    } catch {
      toast.error('复制失败，请手动选择地址')
    }
  }

  return (
    <div className="space-y-4">
      <section className="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-elevated)] p-4 shadow-[var(--shadow-sm)]">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-[11px] font-semibold uppercase tracking-[0.16em] text-[var(--color-text-muted)]">当前文档</p>
            <p className="mt-1 text-sm font-semibold text-[var(--color-text-primary)]">{currentDocLabel}</p>
            <p className="mt-1 text-xs text-[var(--color-text-muted)]">{currentGroupLabel}</p>
          </div>
          <span className="rounded-lg bg-[var(--color-accent-muted)] p-2 text-[var(--color-accent)]">
            <Terminal className="h-4 w-4" />
          </span>
        </div>
      </section>

      <section className="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-elevated)] p-4 shadow-[var(--shadow-sm)]">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Server className="h-4 w-4 text-[var(--color-accent)]" />
            <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">运行状态</h2>
          </div>
          <button
            type="button"
            onClick={onRefresh}
            disabled={launchContextLoading}
            className="rounded-md p-1.5 text-[var(--color-text-muted)] transition hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-text-primary)] disabled:opacity-50"
            title="刷新状态"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${launchContextLoading ? 'animate-spin' : ''}`} />
          </button>
        </div>
        <div className="mt-3 flex items-center gap-2 text-sm">
          <span className={`h-2 w-2 rounded-full ${launchServerReady ? 'bg-[var(--color-success)]' : 'bg-[var(--color-warning)]'}`} />
          <span className="text-[var(--color-text-secondary)]">{launchServerReady ? 'Launch API 在线' : '等待 Launch API'}</span>
        </div>
        <button type="button" onClick={() => void copyBaseUrl()} className="mt-3 flex w-full items-center gap-2 rounded-lg bg-[var(--color-bg-muted)] px-2.5 py-2 text-left font-mono text-[11px] text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]">
          <span className="min-w-0 flex-1 truncate">{launchBaseUrl}</span>
          <Copy className="h-3.5 w-3.5 shrink-0" />
        </button>
        <div className="mt-3 flex flex-wrap gap-1.5 text-[11px]">
          <span className="inline-flex items-center gap-1 rounded-full border border-[var(--color-border-default)] px-2 py-1 text-[var(--color-text-muted)]">
            <KeyRound className="h-3 w-3" />
            {apiAuth.enabled ? 'API Key 已启用' : '本地免鉴权'}
          </span>
          {launchServerReady && <span className="inline-flex items-center gap-1 rounded-full border border-emerald-200 bg-emerald-50 px-2 py-1 text-emerald-700"><CheckCircle2 className="h-3 w-3" />可调用</span>}
        </div>
      </section>

      <section className="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-elevated)] p-4 shadow-[var(--shadow-sm)]">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">本组文档</h2>
          <span className="text-[11px] text-[var(--color-text-muted)]">{relatedDocs.length} 篇</span>
        </div>
        <div className="mt-2 space-y-1">
          {relatedDocs.map((doc) => (
            <button
              key={doc.id}
              type="button"
              onClick={() => onSelectDoc(doc.id)}
              className={`flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-xs transition ${doc.id === currentDocId ? 'bg-[var(--color-accent-muted)] font-medium text-[var(--color-text-primary)]' : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-muted)]'}`}
            >
              <ChevronRight className="h-3.5 w-3.5 shrink-0 text-[var(--color-text-muted)]" />
              <span className="truncate">{doc.label}</span>
            </button>
          ))}
        </div>
      </section>

      <section className="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-muted)] p-4">
        <p className="text-xs font-semibold text-[var(--color-text-primary)]">需要快速接入？</p>
        <p className="mt-1.5 text-xs leading-5 text-[var(--color-text-secondary)]">CLI 与 MCP 都复用同一套 Launch API，适合脚本和 Agent 调用。</p>
        <Button size="sm" variant="secondary" className="mt-3 w-full" onClick={() => onSelectDoc('api-overview')}>
          查看接入方式
          <ExternalLink className="h-3.5 w-3.5" />
        </Button>
      </section>
    </div>
  )
}
