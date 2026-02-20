import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { Plus } from 'lucide-react'
import { MyKnowledgeBasesList } from './-MyKnowledgeBasesList'
import { SharedKnowledgeBasesList } from './-SharedKnowledgeBasesList'
import { PublicKnowledgeBasesList } from './-PublicKnowledgeBasesList'
import { CreateKnowledgeBaseDialog } from './-CreateKnowledgeBaseDialog'

export const Route = createFileRoute('/_authenticated/ai/knowledge-bases/')({
  component: KnowledgeBasesPage,
})

function KnowledgeBasesPage() {
  const [showCreateDialog, setShowCreateDialog] = useState(false)

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">My Knowledge Bases</h1>
        <Button onClick={() => setShowCreateDialog(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Create Knowledge Base
        </Button>
      </div>

      <Tabs defaultValue="my-kbs">
        <TabsList>
          <TabsTrigger value="my-kbs">My Knowledge Bases</TabsTrigger>
          <TabsTrigger value="shared">Shared with Me</TabsTrigger>
          <TabsTrigger value="public">Public</TabsTrigger>
        </TabsList>

        <TabsContent value="my-kbs" className="mt-4">
          <MyKnowledgeBasesList />
        </TabsContent>

        <TabsContent value="shared" className="mt-4">
          <SharedKnowledgeBasesList />
        </TabsContent>

        <TabsContent value="public" className="mt-4">
          <PublicKnowledgeBasesList />
        </TabsContent>
      </Tabs>

      {showCreateDialog && <CreateKnowledgeBaseDialog onClose={() => setShowCreateDialog(false)} />}
    </div>
  )
}
