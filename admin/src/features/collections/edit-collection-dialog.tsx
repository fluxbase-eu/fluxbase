import { useState } from 'react'
import { collectionsApi, type CollectionSummary, type UpdateCollectionRequest, type CollectionMember } from '@/lib/api'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'

interface EditCollectionDialogProps {
  collection: CollectionSummary
  onClose: () => void
  onManageMembers?: () => void // Optional callback to open full members dialog
}

function EditCollectionDialog({ collection, onClose, onManageMembers }: EditCollectionDialogProps) {
  const [name, setName] = useState(collection.name)
  const [description, setDescription] = useState(collection.description || '')
  const [isOpen, setIsOpen] = useState(true) // Track dialog open state for query
  const queryClient = useQueryClient()

  const { data: members = [] } = useQuery({
    queryKey: ['collection-members', collection.id],
    queryFn: () => collectionsApi.listMembers(collection.id),
    enabled: isOpen, // Only fetch when dialog is open
  })

  const updateMutation = useMutation({
    mutationFn: async (data: UpdateCollectionRequest) => {
      return await collectionsApi.update(collection.id, data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collections'] })
      onClose()
    },
  })

  const handleClose = () => {
    setIsOpen(false)
    onClose()
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    const data: UpdateCollectionRequest = {
      name: name.trim(),
      description: description.trim() || undefined,
    }

    updateMutation.mutate(data)
  }

  return (
    <Dialog open onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Edit Collection</DialogTitle>
          <DialogDescription>
            Update the collection details
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-name">Name *</Label>
            <Input
              id="edit-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Collection"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-slug">Slug</Label>
            <Input
              id="edit-slug"
              value={collection.slug}
              disabled
              className="bg-muted"
            />
            <p className="text-xs text-muted-foreground">
              Slug cannot be changed after creation
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-description">Description</Label>
            <Textarea
              id="edit-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="A brief description of this collection"
              rows={3}
            />
          </div>

          {/* Members Section */}
          {members.length > 0 && (
            <div className="space-y-2">
              <Label>Members ({members.length})</Label>
              <div className="border rounded-md p-3 bg-muted/20">
                <ScrollArea className="h-32">
                  <div className="space-y-2 pr-4">
                    {members.map((member: CollectionMember) => (
                      <div
                        key={member.user_id}
                        className="flex items-center justify-between text-sm rounded-md bg-background p-2"
                      >
                        <div className="flex-1 min-w-0">
                          <p className="font-medium truncate">
                            {member.name || member.email || member.user_id}
                          </p>
                          {member.email && member.name && (
                            <p className="text-xs text-muted-foreground truncate">
                              {member.email}
                            </p>
                          )}
                        </div>
                        <Badge variant="secondary" className="capitalize ml-2">
                          {member.role}
                        </Badge>
                      </div>
                    ))}
                  </div>
                </ScrollArea>
              </div>
              {onManageMembers && (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="w-full"
                  onClick={onManageMembers}
                >
                  Manage Members →
                </Button>
              )}
            </div>
          )}

          <div className="flex justify-end gap-2 pt-4">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={updateMutation.isPending || !name.trim()}
            >
              {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export { EditCollectionDialog }
