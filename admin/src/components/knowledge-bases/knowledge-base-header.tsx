import { Link } from '@tanstack/react-router'
import {
  ArrowLeft,
  BookOpen,
  Lock,
  Users,
  Globe,
  type LucideIcon,
} from 'lucide-react'
import type { KnowledgeBaseSummary, KBVisibility } from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  KnowledgeBaseTabs,
  type KnowledgeBaseTab,
} from '@/components/knowledge-bases/knowledge-base-tabs'

interface KnowledgeBaseHeaderProps {
  knowledgeBase: KnowledgeBaseSummary
  activeTab: KnowledgeBaseTab
  /** Optional page-specific action buttons */
  actions?: React.ReactNode
  /** Optional description override */
  description?: string
}

// Tab-specific default descriptions
const TAB_DESCRIPTIONS: Record<KnowledgeBaseTab, string> = {
  documents: 'Manage documents in this knowledge base',
  tables: 'Export database tables as documents',
  graph: 'Visualize entities and relationships extracted from documents',
  search: 'Search documents using semantic similarity',
  settings: 'Configure knowledge base settings',
}

// Visibility badge configuration
const VISIBILITY_CONFIG: Record<
  KBVisibility,
  { icon: LucideIcon; label: string; className: string }
> = {
  private: {
    icon: Lock,
    label: 'Private',
    className: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  },
  shared: {
    icon: Users,
    label: 'Shared',
    className: 'bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300',
  },
  public: {
    icon: Globe,
    label: 'Public',
    className:
      'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300',
  },
}

function VisibilityBadge({ visibility }: { visibility: KBVisibility }) {
  const config = VISIBILITY_CONFIG[visibility] || VISIBILITY_CONFIG.private
  const Icon = config.icon

  return (
    <Badge variant='outline' className={`gap-1 ${config.className}`}>
      <Icon className='h-3 w-3' />
      {config.label}
    </Badge>
  )
}

export function KnowledgeBaseHeader({
  knowledgeBase,
  activeTab,
  actions,
  description,
}: KnowledgeBaseHeaderProps) {
  const displayDescription =
    description || knowledgeBase.description || TAB_DESCRIPTIONS[activeTab]

  return (
    <div className='space-y-4'>
      {/* Back navigation */}
      <div className='flex items-center gap-4'>
        <Button variant='ghost' size='sm' asChild>
          <Link to='/knowledge-bases'>
            <ArrowLeft className='mr-2 h-4 w-4' />
            Back to Knowledge Bases
          </Link>
        </Button>
      </div>

      {/* Main header area */}
      <div className='flex items-start justify-between gap-4'>
        <div className='flex items-start gap-3'>
          {/* Icon container */}
          <div className='bg-primary/10 flex h-12 w-12 shrink-0 items-center justify-center rounded-lg'>
            <BookOpen className='text-primary h-6 w-6' />
          </div>

          {/* Title and description */}
          <div className='space-y-1'>
            <div className='flex flex-wrap items-center gap-2'>
              <h1 className='text-2xl font-bold tracking-tight'>
                {knowledgeBase.name}
              </h1>
              <VisibilityBadge visibility={knowledgeBase.visibility} />
              {!knowledgeBase.enabled && (
                <Badge variant='destructive' className='gap-1'>
                  Disabled
                </Badge>
              )}
            </div>
            <p className='text-muted-foreground max-w-2xl text-sm'>
              {displayDescription}
            </p>
          </div>
        </div>

        {/* Page-specific actions */}
        {actions && (
          <div className='flex shrink-0 items-center gap-2'>{actions}</div>
        )}
      </div>

      {/* Stats row */}
      <div className='flex flex-wrap items-center gap-x-4 gap-y-2 text-sm'>
        <div className='flex items-center gap-1.5'>
          <span className='text-muted-foreground'>Documents:</span>
          <Badge variant='secondary'>{knowledgeBase.document_count}</Badge>
        </div>
        <div className='flex items-center gap-1.5'>
          <span className='text-muted-foreground'>Chunks:</span>
          <Badge variant='secondary'>{knowledgeBase.total_chunks}</Badge>
        </div>
        <div className='flex items-center gap-1.5'>
          <span className='text-muted-foreground'>Model:</span>
          <Badge variant='outline' className='text-xs'>
            {knowledgeBase.embedding_model || 'Default'}
          </Badge>
        </div>
        <div className='flex items-center gap-1.5'>
          <span className='text-muted-foreground'>Namespace:</span>
          <Badge variant='outline' className='font-mono text-xs'>
            {knowledgeBase.namespace}
          </Badge>
        </div>
      </div>

      {/* Tab navigation */}
      <KnowledgeBaseTabs
        activeTab={activeTab}
        knowledgeBaseId={knowledgeBase.id}
      />
    </div>
  )
}
