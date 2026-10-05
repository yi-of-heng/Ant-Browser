import { ArrowRight, Bot, Layers3, Play, Server } from 'lucide-react'

interface LaunchDocsQuickStartProps {
  onOpenDoc: (id: string) => void
}

const steps = [
  { number: '01', title: '确认服务', description: '检查桌面应用与 Launch API 是否可用。', icon: Server, target: 'api-overview' },
  { number: '02', title: '准备实例', description: '选择代理，创建或复制持久化实例。', icon: Layers3, target: 'api-profiles-launch' },
  { number: '03', title: '运行自动化', description: '启动实例，再通过脚本或 CDP 接管。', icon: Play, target: 'api-automation' },
]

export function LaunchDocsQuickStart({ onOpenDoc }: LaunchDocsQuickStartProps) {
  return (
    <div className="space-y-4">
      <section className="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-elevated)] p-5 shadow-[var(--shadow-sm)]">
        <div className="flex items-center gap-2">
          <span className="rounded-lg bg-[var(--color-accent-muted)] p-2 text-[var(--color-accent)]"><Bot className="h-4 w-4" /></span>
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[var(--color-text-muted)]">Quick start</p>
            <h2 className="text-lg font-semibold text-[var(--color-text-primary)]">从配置到自动化，只需三步</h2>
          </div>
        </div>
        <div className="mt-5 grid gap-3 md:grid-cols-3">
          {steps.map((step) => {
            const Icon = step.icon
            return (
              <button
                key={step.number}
                type="button"
                onClick={() => onOpenDoc(step.target)}
                className="group rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-4 text-left transition hover:border-[var(--color-accent)] hover:shadow-[var(--shadow-sm)]"
              >
                <div className="flex items-center justify-between">
                  <span className="font-mono text-xs font-semibold text-[var(--color-accent)]">{step.number}</span>
                  <Icon className="h-4 w-4 text-[var(--color-text-muted)] group-hover:text-[var(--color-accent)]" />
                </div>
                <p className="mt-3 text-sm font-semibold text-[var(--color-text-primary)]">{step.title}</p>
                <p className="mt-1 text-xs leading-5 text-[var(--color-text-secondary)]">{step.description}</p>
                <span className="mt-3 inline-flex items-center gap-1 text-xs font-medium text-[var(--color-accent)]">查看指南 <ArrowRight className="h-3.5 w-3.5" /></span>
              </button>
            )
          })}
        </div>
      </section>
      <section className="grid gap-3 sm:grid-cols-3">
        {[
          { label: 'HTTP API', description: '直接对接本地服务', doc: 'api-overview' },
          { label: 'antctl CLI', description: '批处理与终端脚本', doc: 'agent-cli' },
          { label: 'ant-mcp', description: 'Agent 工具调用', doc: 'agent-mcp' },
        ].map((item) => (
          <button key={item.label} type="button" onClick={() => onOpenDoc(item.doc)} className="flex items-center justify-between rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-elevated)] px-4 py-3 text-left hover:border-[var(--color-accent)]">
            <span><strong className="block text-sm text-[var(--color-text-primary)]">{item.label}</strong><small className="mt-0.5 block text-xs text-[var(--color-text-muted)]">{item.description}</small></span>
            <ArrowRight className="h-4 w-4 shrink-0 text-[var(--color-text-muted)]" />
          </button>
        ))}
      </section>
    </div>
  )
}
