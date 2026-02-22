import { Link } from '@tanstack/react-router'
import { FileText, Database, GitBranch, Search, Settings } from 'lucide-react'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'

export type KnowledgeBaseTab =
  | 'documents'
  | 'tables'
  | 'graph'
  | 'search'
  | 'settings'

interface KnowledgeBaseTabsProps {
  activeTab: KnowledgeBaseTab
  knowledgeBaseId: string
}

export function KnowledgeBaseTabs({
  activeTab,
  knowledgeBaseId,
}: KnowledgeBaseTabsProps) {
  const id = knowledgeBaseId

  return (
    <Tabs value={activeTab} className='w-full'>
      <TabsList className='grid w-full grid-cols-5'>
        <TabsTrigger value='documents' asChild>
          <Link
            to='/knowledge-bases/$id'
            params={{ id }}
            className='flex items-center gap-2'
          >
            <FileText className='h-4 w-4' />
            Documents
          </Link>
        </TabsTrigger>
        <TabsTrigger value='tables' asChild>
          <Link
            to='/knowledge-bases/$id/tables'
            params={{ id }}
            className='flex items-center gap-2'
          >
            <Database className='h-4 w-4' />
            Tables
          </Link>
        </TabsTrigger>
        <TabsTrigger value='graph' asChild>
          <Link
            to='/knowledge-bases/$id/graph'
            params={{ id }}
            className='flex items-center gap-2'
          >
            <GitBranch className='h-4 w-4' />
            Knowledge Graph
          </Link>
        </TabsTrigger>
        <TabsTrigger value='search' asChild>
          <Link
            to='/knowledge-bases/$id/search'
            params={{ id }}
            className='flex items-center gap-2'
          >
            <Search className='h-4 w-4' />
            Search
          </Link>
        </TabsTrigger>
        <TabsTrigger value='settings' asChild>
          <Link
            to='/knowledge-bases/$id/settings'
            params={{ id }}
            className='flex items-center gap-2'
          >
            <Settings className='h-4 w-4' />
            Settings
          </Link>
        </TabsTrigger>
      </TabsList>
    </Tabs>
  )
}
