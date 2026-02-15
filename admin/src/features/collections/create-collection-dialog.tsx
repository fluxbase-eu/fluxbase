import { useState } from 'react'
import { collectionsApi, type CreateCollectionRequest, type CollectionMemberRole } from '@/lib/api'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ChevronDown, ChevronUp, UserPlus, X } from 'lucide-react'

interface CreateCollectionDialogProps {
  onClose: () => void
}

interface PendingMember {
  userId: string
  role: CollectionMemberRole
}

function CreateCollectionDialog({ onClose }: CreateCollectionDialogProps) {
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [description, setDescription] = useState('')
  const [slugManuallyEdited, setSlugManuallyEdited] = useState(false)
  const [membersExpanded, setMembersExpanded] = useState(false)
  const [pendingMembers, setPendingMembers] = useState<PendingMember[]>([])
  const [newMemberUserId, setNewMemberUserId] = useState('')
  const [newMemberRole, setNewMemberRole] = useState<CollectionMemberRole>('viewer')
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: async (data: CreateCollectionRequest) => {
      const result = await collectionsApi.create(data)
      // After collection is created, add pending members
      if (result.id && pendingMembers.length > 0) {
        try {
          await Promise.all(
            pendingMembers.map(member =>
              collectionsApi.addMember(result.id, member.userId, member.role)
            )
          )
        } catch {
          // If member addition fails, still consider the operation successful
          // but show a warning toast
          toast.error('Collection created, but some members could not be added')
        }
      }
      return result
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collections'] })
      toast.success('Collection created successfully')
      onClose()
    },
  })

  // Auto-generate slug from name (only if user hasn't manually edited it)
  const handleNameChange = (value: string) => {
    setName(value)
    if (!slugManuallyEdited) {
      // Generate slug from name: lowercase, replace spaces with hyphens, remove special chars
      const generatedSlug = value
        .toLowerCase()
        .replace(/[^a-z0-9\s-]/g, '')
        .replace(/\s+/g, '-')
        .replace(/-+/g, '-')
        .trim()
      setSlug(generatedSlug)
    }
  }

  const handleSlugChange = (value: string) => {
    setSlug(value)
    setSlugManuallyEdited(true) // Stop auto-generating once user manually edits
  }

  const handleAddPendingMember = () => {
    if (!newMemberUserId.trim()) return

    // Check if member already exists
    if (pendingMembers.some(m => m.userId === newMemberUserId.trim())) {
      toast.error('Member already added')
      return
    }

    setPendingMembers([...pendingMembers, {
      userId: newMemberUserId.trim(),
      role: newMemberRole,
    }])
    setNewMemberUserId('')
    setNewMemberRole('viewer')
  }

  const handleRemovePendingMember = (userId: string) => {
    setPendingMembers(pendingMembers.filter(m => m.userId !== userId))
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    const data: CreateCollectionRequest = {
      name: name.trim(),
      slug: slug.trim() || undefined,
      description: description.trim() || undefined,
    }

    createMutation.mutate(data)
  }

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Create Collection</DialogTitle>
          <DialogDescription>
            Create a new collection to organize your knowledge bases
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder="My Collection"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="slug">Slug</Label>
            <Input
              id="slug"
              value={slug}
              onChange={(e) => handleSlugChange(e.target.value)}
              placeholder="my-collection"
            />
            <p className="text-xs text-muted-foreground">
              URL-friendly identifier. Auto-generated from name if left blank.
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="A brief description of this collection"
              rows={3}
            />
          </div>

          {/* Collapsible Members Section */}
          <div className="space-y-2">
            <Button
              type="button"
              variant="outline"
              className="w-full justify-between"
              onClick={() => setMembersExpanded(!membersExpanded)}
            >
              <span>Add Members</span>
              {membersExpanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
            </Button>

            {membersExpanded && (
              <div className="space-y-3 border rounded-md p-3 bg-muted/20">
                {/* Add Member Form */}
                <form onSubmit={(e) => { e.preventDefault(); handleAddPendingMember(); }} className="flex gap-2">
                  <div className="flex-1">
                    <Label htmlFor="member-user-id" className="sr-only">
                      User ID or Email
                    </Label>
                    <Input
                      id="member-user-id"
                      placeholder="Enter user ID or email"
                      value={newMemberUserId}
                      onChange={(e) => setNewMemberUserId(e.target.value)}
                      disabled={createMutation.isPending}
                    />
                  </div>
                  <div className="w-32">
                    <Label htmlFor="member-role" className="sr-only">
                      Role
                    </Label>
                    <Select value={newMemberRole} onValueChange={(v: CollectionMemberRole) => setNewMemberRole(v)}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="viewer">Viewer</SelectItem>
                        <SelectItem value="editor">Editor</SelectItem>
                        <SelectItem value="owner">Owner</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <Button
                    type="submit"
                    disabled={!newMemberUserId.trim() || createMutation.isPending}
                    size="icon"
                  >
                    <UserPlus className="h-4 w-4" />
                  </Button>
                </form>

                {/* Pending Members List */}
                {pendingMembers.length > 0 && (
                  <div className="space-y-2">
                    <Label className="text-xs text-muted-foreground">
                      Members to be added ({pendingMembers.length})
                    </Label>
                    <div className="space-y-1">
                      {pendingMembers.map((member) => (
                        <div
                          key={member.userId}
                          className="flex items-center justify-between rounded-md border bg-background p-2 text-sm"
                        >
                          <span className="flex-1 truncate">{member.userId}</span>
                          <div className="flex items-center gap-2">
                            <Badge variant="secondary" className="capitalize">
                              {member.role}
                            </Badge>
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              className="h-6 w-6 text-destructive hover:text-destructive"
                              onClick={() => handleRemovePendingMember(member.userId)}
                              disabled={createMutation.isPending}
                            >
                              <X className="h-3 w-3" />
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Members will be added after the collection is created
                    </p>
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={createMutation.isPending || !name.trim()}
            >
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export { CreateCollectionDialog }
