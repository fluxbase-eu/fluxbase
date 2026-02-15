import { useState } from 'react'
import { X, Users } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { collectionsApi, knowledgeBasesApi } from '@/lib/api'
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from '@/components/ui/select'
import { CollectionMembersDialog } from './collection-members-dialog'

interface CollectionDetailPanelProps {
  collectionId: string
  onClose: () => void
}

function CollectionDetailPanel({ collectionId, onClose }: CollectionDetailPanelProps) {
  const queryClient = useQueryClient()
  const [selectedKbId, setSelectedKbId] = useState<string>('')
  const [membersDialogOpen, setMembersDialogOpen] = useState(false)

  const { data: collection, isLoading: collectionLoading } = useQuery({
    queryKey: ['collection', collectionId],
    queryFn: () => collectionsApi.get(collectionId),
  })

  const { data: knowledgeBases = [], isLoading: kbsLoading } = useQuery({
    queryKey: ['collection-kbs', collectionId],
    queryFn: () => collectionsApi.listKnowledgeBases(collectionId),
  })

  const { data: allKnowledgeBases = [] } = useQuery({
    queryKey: ['my-knowledge-bases'],
    queryFn: () => knowledgeBasesApi.list(),
  })

  const addMutation = useMutation({
    mutationFn: async (kbId: string) => {
      await collectionsApi.addKnowledgeBase(collectionId, kbId)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collection-kbs', collectionId] })
      queryClient.invalidateQueries({ queryKey: ['collections'] })
      setSelectedKbId('')
    },
  })

  const removeMutation = useMutation({
    mutationFn: async (kbId: string) => {
      await collectionsApi.removeKnowledgeBase(collectionId, kbId)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collection-kbs', collectionId] })
      queryClient.invalidateQueries({ queryKey: ['collections'] })
    },
  })

  // Knowledge bases not in this collection
  const availableKbs = allKnowledgeBases.filter(
    (kb) => !knowledgeBases.some((k) => k.id === kb.id)
  )

  const canManageKBs =
    collection?.current_user_role === 'owner' || collection?.current_user_role === 'editor'

  const canManageMembers = collection?.current_user_role === 'owner'

  if (collectionLoading) {
    return (
      <div className="h-full flex items-center justify-center">
        <p className="text-muted-foreground">Loading collection...</p>
      </div>
    )
  }

  if (!collection) {
    return (
      <div className="h-full flex items-center justify-center">
        <p className="text-destructive">Collection not found</p>
      </div>
    )
  }

  return (
    <>
      <div className="h-full flex flex-col bg-background border-l">
        <div className="flex items-center justify-between p-4 border-b">
          <div>
            <h2 className="text-lg font-semibold">{collection.name}</h2>
            {collection.description && (
              <p className="text-sm text-muted-foreground mt-1">{collection.description}</p>
            )}
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="w-4 h-4" />
          </Button>
        </div>

        {/* Stats Bar */}
        <div className="flex items-center gap-4 px-4 py-3 border-b bg-muted/30">
          <div className="flex items-center gap-2 text-sm">
            <span className="text-muted-foreground">Role:</span>
            <Badge variant="outline" className="capitalize">
              {collection.current_user_role || 'none'}
            </Badge>
          </div>
          <div className="flex items-center gap-2 text-sm">
            <span className="text-muted-foreground">Members:</span>
            <span className="font-medium">{collection.member_count || 0}</span>
          </div>
          {canManageMembers && (
            <Button
              variant="ghost"
              size="sm"
              className="ml-auto"
              onClick={() => setMembersDialogOpen(true)}
            >
              <Users className="w-4 h-4 mr-2" />
              Manage
            </Button>
          )}
        </div>

        <div className="flex-1 overflow-hidden flex flex-col">
          <div className="p-4 border-b">
            <div className="flex items-center gap-2 mb-4">
              <span className="text-sm font-medium">Knowledge Bases</span>
              <Badge variant="secondary">{knowledgeBases.length}</Badge>
            </div>

            {canManageKBs && availableKbs.length > 0 && (
              <div className="flex gap-2">
                <Select value={selectedKbId} onValueChange={setSelectedKbId}>
                  <SelectTrigger className="flex-1">
                    <SelectValue placeholder="Select a knowledge base to add" />
                  </SelectTrigger>
                  <SelectContent>
                    {availableKbs.map((kb) => (
                      <SelectItem key={kb.id} value={kb.id}>
                        {kb.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button
                  onClick={() => selectedKbId && addMutation.mutate(selectedKbId)}
                  disabled={!selectedKbId || addMutation.isPending}
                  size="sm"
                >
                  {addMutation.isPending ? 'Adding...' : 'Add'}
                </Button>
              </div>
            )}

            {!canManageKBs && (
              <p className="text-xs text-muted-foreground">
                You need editor or owner permissions to add knowledge bases to this collection.
              </p>
            )}
          </div>

          <ScrollArea className="flex-1">
            <div className="p-4 space-y-2">
              {kbsLoading ? (
                <p className="text-sm text-muted-foreground">Loading knowledge bases...</p>
              ) : knowledgeBases.length === 0 ? (
                <p className="text-sm text-muted-foreground text-center py-8">
                  No knowledge bases in this collection yet
                </p>
              ) : (
                knowledgeBases.map((kb) => (
                  <div
                    key={kb.id}
                    className="flex items-center justify-between p-3 border rounded-lg hover:bg-muted/50 transition-colors"
                  >
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{kb.name}</p>
                      {kb.description && (
                        <p className="text-xs text-muted-foreground truncate">{kb.description}</p>
                      )}
                    </div>
                    {canManageKBs && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => removeMutation.mutate(kb.id)}
                        disabled={removeMutation.isPending}
                        className="text-destructive hover:text-destructive"
                      >
                        {removeMutation.isPending ? 'Removing...' : 'Remove'}
                      </Button>
                    )}
                  </div>
                ))
              )}
            </div>
          </ScrollArea>
        </div>
      </div>

      {membersDialogOpen && (
        <CollectionMembersDialog
          collectionId={collectionId}
          collectionName={collection.name}
          onClose={() => setMembersDialogOpen(false)}
        />
      )}
    </>
  )
}

export { CollectionDetailPanel }
