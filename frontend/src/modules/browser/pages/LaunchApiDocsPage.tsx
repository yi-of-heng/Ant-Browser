import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useLaunchContext } from '../hooks/useLaunchContext'
import {
  DOC_GROUPS,
  findDocById,
  findGroupByDocId,
  getDefaultDoc,
  getAdjacentDocs,
  renderDocWithLaunchContext,
} from './launchApiDocs/catalog'
import { LaunchDocsFlowPage } from './launchApiDocs/LaunchDocsFlowPage'
import { LaunchDocsLayout } from './launchApiDocs/LaunchDocsLayout'
import { LaunchDocsMarkdownContent } from './launchApiDocs/LaunchDocsMarkdownContent'
import { LaunchDocsPager } from './launchApiDocs/LaunchDocsPager'
import { LaunchDocsSidebar } from './launchApiDocs/LaunchDocsSidebar'
import { LaunchDocsHeader } from './launchApiDocs/LaunchDocsHeader'
import { LaunchDocsContextRail } from './launchApiDocs/LaunchDocsContextRail'
import { LaunchDocsQuickStart } from './launchApiDocs/LaunchDocsQuickStart'
import { StructuredApiDocsPage } from './launchApiDocs/StructuredApiDocsPage'
import {
  getStructuredApiParentDocId,
  isStructuredApiDocId,
  isStructuredApiEndpointDocId,
  type StructuredApiDocId,
} from './launchApiDocs/structuredApiDocs'

export function LaunchApiDocsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const firstDoc = getDefaultDoc()
  const [activeId, setActiveId] = useState(firstDoc.id)
  const {
    launchBaseUrl,
    launchServerReady,
    launchContextLoading,
    apiAuth,
    refreshLaunchContext,
  } = useLaunchContext()

  const activeDoc = findDocById(activeId) || firstDoc
  const activeGroup = findGroupByDocId(activeDoc.id) || DOC_GROUPS.find((group) => group.id === 'api') || DOC_GROUPS[0]
  const { previous, next } = isStructuredApiEndpointDocId(activeDoc.id) ? { previous: null, next: null } : getAdjacentDocs(activeDoc.id)
  const sidebarActiveId = isStructuredApiDocId(activeDoc.id) ? getStructuredApiParentDocId(activeDoc.id) : activeDoc.id

  const selectDoc = (id: string, syncURL: boolean) => {
    const doc = findDocById(id)
    if (!doc) {
      return false
    }

    setActiveId(doc.id)
    if (syncURL) {
      setSearchParams({ doc: doc.id })
    }
    return true
  }

  useEffect(() => {
    const requestedDoc = searchParams.get('doc')?.trim() || ''
    if (!requestedDoc || requestedDoc === activeId) {
      return
    }

    if (!selectDoc(requestedDoc, false)) {
      setSearchParams({ doc: firstDoc.id })
    }
  }, [activeId, firstDoc.id, searchParams, setSearchParams])

  const renderedContent = renderDocWithLaunchContext(activeDoc.content, launchBaseUrl, apiAuth.header)
    .replace(/^# [^\n]+\n+/, '')

  return (
    <LaunchDocsLayout
      sidebar={(
        <LaunchDocsSidebar
          groups={DOC_GROUPS}
          activeId={sidebarActiveId}
          onSelect={(id) => {
            void selectDoc(id, true)
          }}
        />
      )}
      header={(
        <LaunchDocsHeader
          activeGroupLabel={activeGroup.label}
          activeDocLabel={activeDoc.label}
          activeDocSummary={activeDoc.summary}
          launchBaseUrl={launchBaseUrl}
          launchServerReady={launchServerReady}
          apiAuthEnabled={apiAuth.enabled}
          onBack={() => navigate('/browser/list')}
          onJumpTutorial={() => void selectDoc('tutorial-basic', true)}
          onJumpCoreIntro={() => void selectDoc('core-management', true)}
          onJumpProxyIntro={() => void selectDoc('proxy-usage', true)}
          onJumpApiOverview={() => void selectDoc('api-overview', true)}
        />
      )}
      contextRail={(
        <LaunchDocsContextRail
          currentGroupLabel={activeGroup.label}
          currentDocId={activeDoc.id}
          currentDocLabel={activeDoc.label}
          launchBaseUrl={launchBaseUrl}
          launchServerReady={launchServerReady}
          launchContextLoading={launchContextLoading}
          apiAuth={apiAuth}
          relatedDocs={activeGroup.items}
          onRefresh={() => {
            void refreshLaunchContext(true)
          }}
          onSelectDoc={(id) => {
            void selectDoc(id, true)
          }}
        />
      )}
      content={(
        <div className="space-y-4">
          {activeDoc.id === 'tutorial-basic' && (
            <LaunchDocsQuickStart onOpenDoc={(id) => { void selectDoc(id, true) }} />
          )}
          {activeDoc.id === 'tutorial-flow'
            ? <LaunchDocsFlowPage baseUrl={launchBaseUrl} />
            : isStructuredApiDocId(activeDoc.id)
              ? (
                <StructuredApiDocsPage
                  docId={activeDoc.id as StructuredApiDocId}
                  launchBaseUrl={launchBaseUrl}
                  authHeader={apiAuth.header}
                  onOpenDoc={(id) => {
                    void selectDoc(id, true)
                  }}
                />
              )
              : <LaunchDocsMarkdownContent content={renderedContent} docId={activeDoc.id} />}
          <LaunchDocsPager
            previous={previous}
            next={next}
            onSelect={(id) => {
              void selectDoc(id, true)
            }}
          />
        </div>
      )}
    />
  )
}
