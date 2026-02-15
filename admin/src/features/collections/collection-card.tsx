import { useState } from 'react'
import { collectionsApi, type CollectionSummary } from '@/lib/api'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { MoreVertical, Eye, Edit, Trash2, Database, Users } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu'
import { EditCollectionDialog } from './edit-collection-dialog'
import { CollectionDetailPanel } from './collection-detail-panel'
import { CollectionMembersDialog } from './collection-members-dialog'

interface CollectionCardProps {
  collection: CollectionSummary
}

function CollectionCard({ collection }: CollectionCardProps) {
  const queryClient = useQueryClient()
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [membersDialogOpen, setMembersDialogOpen] = useState(false)
  const [detailPanelOpen, setDetailPanelOpen] = useState(false)

  const deleteMutation = useMutation({
    mutationFn: async () => {
      await collectionsApi.delete(collection.id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collections'] })
    },
  })

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this collection? This will not delete the knowledge bases within it.')) {
      return
    }

    try {
      await deleteMutation.mutateAsync()
    } catch (error) {
      // eslint-disable-next-line no-console
      console.error('Error deleting collection:', error)
      alert('Failed to delete collection')
    }
  }

  const canEdit = collection.current_user_role === 'owner'
  const canManageMembers = collection.current_user_role === 'owner'

  return (
    <Card className="hover:shadow-md transition-shadow">
      <CardHeader>
        <div className="flex justify-between items-start">
          <div className="flex-1">
            <CardTitle className="text-lg flex items-center gap-2 flex-wrap">
              {collection.name}
              {collection.current_user_role && (
                <Badge variant="outline" className="text-xs capitalize">
                  {collection.current_user_role}
                </Badge>
              )}
            </CardTitle>
            {collection.description && (
              <p className="text-sm text-muted-foreground mt-2">{collection.description}</p>
            )}
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm">
                <MoreVertical className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => setDetailPanelOpen(true)}>
                <Eye className="w-4 h-4 mr-2" />
                View Details
              </DropdownMenuItem>
              {canManageMembers && (
                <DropdownMenuItem onClick={() => setMembersDialogOpen(true)}>
                  <Users className="w-4 h-4 mr-2" />
                  Manage Members
                </DropdownMenuItem>
              )}
              {canEdit && (
                <DropdownMenuItem onClick={() => setEditDialogOpen(true)}>
                  <Edit className="w-4 h-4 mr-2" />
                  Edit
                </DropdownMenuItem>
              )}
              {canEdit && (
                <DropdownMenuItem
                  className="text-destructive"
                  onClick={handleDelete}
                  disabled={deleteMutation.isPending}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          <div className="flex items-center gap-2">
            <Database className="w-4 h-4" />
            <span>{collection.kb_count || 0} KBs</span>
          </div>
          <div className="flex items-center gap-2">
            <Users className="w-4 h-4" />
            <span>{collection.member_count || 0} members</span>
          </div>
        </div>
      </CardContent>
      <CardFooter className="flex gap-2">
        <Button
          variant="outline"
          className="flex-1"
          onClick={() => setDetailPanelOpen(true)}
        >
          <Eye className="w-4 h-4 mr-2" />
          View
        </Button>
        {canManageMembers && (
          <Button
            variant="outline"
            className="flex-1"
            onClick={() => setMembersDialogOpen(true)}
          >
            <Users className="w-4 h-4 mr-2" />
            Members
          </Button>
        )}
      </CardFooter>
      {editDialogOpen && (
        <EditCollectionDialog
          collection={collection}
          onClose={() => setEditDialogOpen(false)}
          onManageMembers={() => {
            setEditDialogOpen(false)
            setMembersDialogOpen(true)
          }}
        />
      )}
      {membersDialogOpen && (
        <CollectionMembersDialog
          collectionId={collection.id}
          collectionName={collection.name}
          onClose={() => setMembersDialogOpen(false)}
        />
      )}
      {detailPanelOpen && (
        <div className="fixed inset-0 z-50 flex justify-end">
          <div
            className="absolute inset-0 bg-black/50"
            onClick={() => setDetailPanelOpen(false)}
          />
          <div className="relative w-full max-w-md h-full">
            <CollectionDetailPanel
              collectionId={collection.id}
              onClose={() => setDetailPanelOpen(false)}
            />
          </div>
        </div>
      )}
    </Card>
  )
}

export { CollectionCard }
