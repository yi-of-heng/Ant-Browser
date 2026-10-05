import { AlertCircle, ArrowLeft, CheckCircle2, ExternalLink, Terminal } from 'lucide-react'
import { Button } from '../../../../shared/components'
import { BrowserOpenURL } from '../../../../wailsjs/runtime/runtime'

interface LaunchDocsHeaderProps {
  activeGroupLabel: string
  activeDocLabel: string
  activeDocSummary: string
  launchBaseUrl: string
  launchServerReady: boolean
  apiAuthEnabled: boolean
  onBack: () => void
  onJumpTutorial: () => void
  onJumpCoreIntro: () => void
  onJumpProxyIntro: () => void
  onJumpApiOverview: () => void
}

export function LaunchDocsHeader({
  activeGroupLabel,
  activeDocLabel,
  activeDocSummary,
  launchBaseUrl,
  launchServerReady,
  apiAuthEnabled,
  onBack,
  onJumpTutorial,
  onJumpCoreIntro,
  onJumpProxyIntro,
  onJumpApiOverview,
}: LaunchDocsHeaderProps) {
  const quickLinks = [
    { label: '使用教程', onClick: onJumpTutorial },
    { label: '内核介绍', onClick: onJumpCoreIntro },
    { label: '代理介绍', onClick: onJumpProxyIntro },
    { label: '接口总览', onClick: onJumpApiOverview },
  ]

  return (
    <section className="overflow-hidden rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-elevated)] shadow-[var(--shadow-sm)]">
      <div className="bg-gradient-to-br from-[var(--color-accent-muted)] via-[var(--color-bg-elevated)] to-[var(--color-bg-elevated)] px-5 py-5 md:px-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.16em] text-[var(--color-text-muted)]">
            <span>Ant Browser</span>
            <span className="text-[var(--color-border-strong)]">/</span>
            <span>{activeGroupLabel}</span>
          </div>
          <h1 className="mt-2 text-2xl font-semibold tracking-tight text-[var(--color-text-primary)] md:text-3xl">{activeDocLabel}</h1>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--color-text-secondary)]">{activeDocSummary}</p>
        </div>

          <div className="flex shrink-0 flex-wrap gap-2">
            <Button size="sm" variant="secondary" onClick={onBack}>
              <ArrowLeft className="h-4 w-4" />
              实例列表
            </Button>
            <Button size="sm" variant="ghost" onClick={() => BrowserOpenURL('https://github.com/yi-of-heng/Ant-Browser')}>
              <ExternalLink className="h-4 w-4" />
              GitHub
            </Button>
          </div>
        </div>

        <div className="mt-5 flex flex-wrap items-center gap-2 text-xs">
          <span className="inline-flex items-center gap-1.5 rounded-full border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-2.5 py-1.5 text-[var(--color-text-secondary)]">
            {launchServerReady
              ? <CheckCircle2 className="h-3.5 w-3.5 text-[var(--color-success)]" />
              : <AlertCircle className="h-3.5 w-3.5 text-[var(--color-warning)]" />}
            {launchServerReady ? 'Launch API 在线' : 'Launch API 未连接'}
          </span>
          <span className="inline-flex max-w-full items-center gap-1.5 rounded-full border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-2.5 py-1.5 font-mono text-[var(--color-text-muted)]">
            <Terminal className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{launchBaseUrl}</span>
          </span>
          <span className="rounded-full border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-2.5 py-1.5 text-[var(--color-text-muted)]">
            {apiAuthEnabled ? 'API Key 已启用' : '本地免鉴权'}
          </span>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-2 border-t border-[var(--color-border-muted)] px-5 py-3 md:px-6">
        <span className="mr-1 text-xs font-medium text-[var(--color-text-muted)]">快速跳转</span>
        {quickLinks.map((link) => (
          <button
            key={link.label}
            onClick={link.onClick}
            className="rounded-full border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-3 py-1.5 text-xs text-[var(--color-text-secondary)] transition-colors hover:border-[var(--color-accent)] hover:text-[var(--color-text-primary)]"
          >
            {link.label}
          </button>
        ))}
      </div>
    </section>
  )
}
