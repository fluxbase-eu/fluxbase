import { useState } from 'react'
import { collectionsApi, type CollectionMember, type CollectionMemberRole } from '@/lib/api'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
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
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from '@/components/ui/select'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { UserPlus, Trash2, Shield, ShieldAlert, ShieldCheck, Loader2 } from 'lucide-react'

interface CollectionMembersDialogProps {
  collectionId: string
  collectionName: string
  onClose: () => void
}

function CollectionMembersDialog({
  collectionId,
  collectionName,
  onClose,
}: CollectionMembersDialogProps) {
  const queryClient = useQueryClient()
  const [newMemberUserId, setNewMemberUserId] = useState('')
  const [newMemberRole, setNewMemberRole] = useState<CollectionMemberRole>('viewer')
  const [memberToRemove, setMemberToRemove] = useState<CollectionMember | null>(null)
  const [memberToUpdate, setMemberToUpdate] = useState<{ member: CollectionMember; newRole: CollectionMemberRole } | null>(null)

  const { data: members = [], isLoading } = useQuery({
    queryKey: ['collection-members', collectionId],
    queryFn: () => collectionsApi.listMembers(collectionId),
  })

  const addMemberMutation = useMutation({
    mutationFn: async () => {
      if (!newMemberUserId.trim()) return
      await collectionsApi.addMember(collectionId, newMemberUserId.trim(), newMemberRole)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collection-members', collectionId] })
      queryClient.invalidateQueries({ queryKey: ['collections'] })
      setNewMemberUserId('')
      setNewMemberRole('viewer')
    },
  })

  const removeMemberMutation = useMutation({
    mutationFn: async (userId: string) => {
      await collectionsApi.removeMember(collectionId, userId)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collection-members', collectionId] })
      queryClient.invalidateQueries({ queryKey: ['collections'] })
      setMemberToRemove(null)
    },
  })

  const updateRoleMutation = useMutation({
    mutationFn: async ({ userId, role }: { userId: string; role: CollectionMemberRole }) => {
      await collectionsApi.updateMemberRole(collectionId, userId, role)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collection-members', collectionId] })
      queryClient.invalidateQueries({ queryKey: ['collections'] })
      setMemberToUpdate(null)
    },
  })

  const handleAddMember = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newMemberUserId.trim()) return
    addMemberMutation.mutate()
  }

  const getRoleIcon = (role: CollectionMemberRole) => {
    switch (role) {
      case 'owner':
        return <ShieldCheck className="w-4 h-4 text-yellow-600 dark:text-yellow-500" />
      case 'editor':
        return <Shield className="w-4 h-4 text-blue-600 dark:text-blue-500" />
      case 'viewer':
        return <ShieldAlert className="w-4 h-4 text-gray-600 dark:text-gray-500" />
    }
  }

  const getRoleBadgeVariant = (role: CollectionMemberRole): 'default' | 'secondary' | 'outline' | 'destructive' => {
    switch (role) {
      case 'owner':
        return 'default'
      case 'editor':
        return 'secondary'
      case 'viewer':
        return 'outline'
    }
  }

  return (
    <>
      <Dialog open onOpenChange={onClose}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Manage Members</DialogTitle>
            <DialogDescription>
              Manage access to collection &quot;{collectionName}&quot;
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {/* Add Member Form */}
            <form onSubmit={handleAddMember} className="flex gap-2">
              <div className="flex-1">
                <Label htmlFor="user-id" className="sr-only">
                  User ID or Email
                </Label>
                <Input
                  id="user-id"
                  placeholder="Enter user ID or email"
                  value={newMemberUserId}
                  onChange={(e) => setNewMemberUserId(e.target.value)}
                  disabled={addMemberMutation.isPending}
                />
              </div>
              <div className="w-32">
                <Label htmlFor="role" className="sr-only">
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
                disabled={!newMemberUserId.trim() || addMemberMutation.isPending}
              >
                {addMemberMutation.isPending ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <UserPlus className="w-4 h-4" />
                )}
              </Button>
            </form>

            {/* Members List */}
            <div className="border rounded-lg">
              <ScrollArea className="h-80">
                {isLoading ? (
                  <div className="flex items-center justify-center py-12">
                    <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
                  </div>
                ) : members.length === 0 ? (
                  <div className="flex flex-col items-center justify-center py-12 text-center">
                    <ShieldAlert className="w-12 h-12 text-muted-foreground mb-4" />
                    <p className="text-sm text-muted-foreground">No members yet</p>
                    <p className="text-xs text-muted-foreground mt-1">
                      Add members to collaborate on this collection
                    </p>
                  </div>
                ) : (
                  <div className="divide-y">
                    {members.map((member) => (
                      <div
                        key={member.user_id}
                        className="flex items-center justify-between p-4 hover:bg-muted/50 transition-colors"
                      >
                        <div className="flex items-center gap-3 flex-1 min-w-0">
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium truncate">
                              {member.name || member.email || member.user_id}
                            </p>
                            {member.email && member.name && (
                              <p className="text-xs text-muted-foreground truncate">{member.email}</p>
                            )}
                          </div>
                          <Badge variant={getRoleBadgeVariant(member.role)} className="capitalize flex items-center gap-1">
                            {getRoleIcon(member.role)}
                            {member.role}
                          </Badge>
                        </div>
                        <div className="flex items-center gap-2 ml-4">
                          {/* Role Change */}
                          {member.role !== 'owner' && (
                            <Select
                              value={member.role}
                              onValueChange={(newRole: CollectionMemberRole) =>
                                setMemberToUpdate({ member, newRole })
                              }
                            >
                              <SelectTrigger className="w-24 h-8">
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent>
                                <SelectItem value="viewer">Viewer</SelectItem>
                                <SelectItem value="editor">Editor</SelectItem>
                                <SelectItem value="owner">Owner</SelectItem>
                              </SelectContent>
                            </Select>
                          )}
                          {/* Remove Member */}
                          {member.role !== 'owner' && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setMemberToRemove(member)}
                              disabled={removeMemberMutation.isPending}
                              className="text-destructive hover:text-destructive hover:bg-destructive/10"
                            >
                              <Trash2 className="w-4 h-4" />
                            </Button>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </ScrollArea>
            </div>

            {/* Legend */}
            <div className="flex flex-wrap gap-4 text-xs text-muted-foreground">
              <div className="flex items-center gap-1">
                <ShieldCheck className="w-3 h-3" />
                <span>Owner: Full access</span>
              </div>
              <div className="flex items-center gap-1">
                <Shield className="w-3 h-3" />
                <span>Editor: Can modify KBs</span>
              </div>
              <div className="flex items-center gap-1">
                <ShieldAlert className="w-3 h-3" />
                <span>Viewer: Read-only</span>
              </div>
            </div>
          </div>

          <div className="flex justify-end pt-4">
            <Button variant="outline" onClick={onClose}>
              Close
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Remove Member Confirmation */}
      <AlertDialog open={!!memberToRemove} onOpenChange={() => setMemberToRemove(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Member?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to remove &quot;{memberToRemove?.name || memberToRemove?.email || memberToRemove?.user_id}&quot;
              from this collection? They will lose access to all knowledge bases in this collection.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => memberToRemove && removeMemberMutation.mutate(memberToRemove.user_id)}
              disabled={removeMemberMutation.isPending}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {removeMemberMutation.isPending ? 'Removing...' : 'Remove'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Update Member Role Confirmation */}
      <AlertDialog open={!!memberToUpdate} onOpenChange={() => setMemberToUpdate(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Change Member Role?</AlertDialogTitle>
            <AlertDialogDescription>
              Change &quot;{memberToUpdate?.member.name || memberToUpdate?.member.email || memberToUpdate?.member.user_id}&quot;
              from &quot;{memberToUpdate?.member.role}&quot; to &quot;{memberToUpdate?.newRole}&quot;?
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() =>
                memberToUpdate &&
                updateRoleMutation.mutate({
                  userId: memberToUpdate.member.user_id,
                  role: memberToUpdate.newRole,
                })
              }
              disabled={updateRoleMutation.isPending}
            >
              {updateRoleMutation.isPending ? 'Updating...' : 'Update Role'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}

export { CollectionMembersDialog }
