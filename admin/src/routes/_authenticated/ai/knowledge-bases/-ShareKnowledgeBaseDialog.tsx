import { useEffect, useMemo } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from '@/components/ui/select'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

interface Permission {
  user_id: string
  permission: string
}

interface User {
  id: string
  email: string
  name?: string
}

interface ShareKnowledgeBaseDialogProps {
  kbId: string
  kbName: string
  open: boolean
  onClose: () => void
}

function ShareKnowledgeBaseDialog({ kbId, kbName, open, onClose }: ShareKnowledgeBaseDialogProps) {
  const queryClient = useQueryClient()

  // Fetch users to share with
  const { data: usersData } = useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const res = await fetch('/api/v1/admin/users')
      if (!res.ok) throw new Error('Failed to fetch users')
      return res.json()
    },
  })

  // Fetch existing permissions
  const { data: permissionsData } = useQuery({
    queryKey: ['kb-permissions', kbId],
    queryFn: async () => {
      const res = await fetch(`/api/v1/ai/knowledge-bases/${kbId}/permissions`)
      if (!res.ok) throw new Error('Failed to fetch permissions')
      return res.json()
    },
  })

  // Compute permissions from data using useMemo
  const permissions = useMemo<Record<string, string>>(() => {
    if (!permissionsData?.permissions) return {}
    const perms: Record<string, string> = {}
    permissionsData.permissions.forEach((p: Permission) => {
      perms[p.user_id] = p.permission
    })
    return perms
  }, [permissionsData])

  const grantMutation = useMutation({
    mutationFn: async ({ userId, permission }: { userId: string; permission: string }) => {
      const res = await fetch(`/api/v1/ai/knowledge-bases/${kbId}/share`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_id: userId, permission }),
      })
      if (!res.ok) throw new Error('Failed to grant permission')
      return res.json()
    },
  })

  const revokeMutation = useMutation({
    mutationFn: async (userId: string) => {
      const res = await fetch(`/api/v1/ai/knowledge-bases/${kbId}/permissions/${userId}`, {
        method: 'DELETE',
      })
      if (!res.ok) throw new Error('Failed to revoke permission')
    },
  })

  // Invalidate queries when mutations succeed
  useEffect(() => {
    if (grantMutation.isSuccess) {
      queryClient.invalidateQueries({ queryKey: ['kb-permissions', kbId] })
    }
  }, [grantMutation.isSuccess, queryClient, kbId])

  useEffect(() => {
    if (revokeMutation.isSuccess) {
      queryClient.invalidateQueries({ queryKey: ['kb-permissions', kbId] })
    }
  }, [revokeMutation.isSuccess, queryClient, kbId])

  const handlePermissionChange = (userId: string, permission: string) => {
    if (permission === 'none') {
      revokeMutation.mutate(userId)
    } else {
      grantMutation.mutate({ userId, permission })
    }
  }

  const users = usersData?.users || []

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Share Knowledge Base</DialogTitle>
          <DialogDescription>Grant access to "{kbName}"</DialogDescription>
        </DialogHeader>

        <div className="space-y-4 max-h-96 overflow-y-auto">
          {users.map((user: User) => (
            <div key={user.id} className="flex items-center justify-between py-2 border-b">
              <div className="flex-1">
                <div className="font-medium">{user.email}</div>
                {user.name && <div className="text-sm text-muted-foreground">{user.name}</div>}
              </div>
              <Select
                value={permissions[user.id] || 'none'}
                onValueChange={(value) => handlePermissionChange(user.id, value)}
              >
                <SelectTrigger className="w-40">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">No Access</SelectItem>
                  <SelectItem value="viewer">Viewer</SelectItem>
                  <SelectItem value="editor">Editor</SelectItem>
                </SelectContent>
              </Select>
            </div>
          ))}
          {users.length === 0 && (
            <div className="text-center py-8 text-muted-foreground">
              No users available to share with.
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

export { ShareKnowledgeBaseDialog }
