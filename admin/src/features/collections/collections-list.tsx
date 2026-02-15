import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { collectionsApi } from '@/lib/api'
import { CollectionCard } from './collection-card'
import { CreateCollectionDialog } from './create-collection-dialog'
import { CollectionDetailPanel } from './collection-detail-panel'
import { Button } from '@/components/ui/button'
import { Plus, AlertCircle } from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface CollectionsListProps {
  showCreateButton?: boolean
}

function CollectionsList({ showCreateButton = true }: CollectionsListProps) {
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [selectedCollectionId, setSelectedCollectionId] = useState<string | null>(null)

  const { data: collections = [], isLoading, error } = useQuery({
    queryKey: ['collections'],
    queryFn: () => collectionsApi.list(),
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">Loading collections...</p>
      </div>
    )
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load collections: {(error as Error).message}
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <div className="relative">
      {showCreateButton && (
        <div className="flex justify-end mb-6">
          <Button onClick={() => setShowCreateDialog(true)}>
            <Plus className="w-4 h-4 mr-2" />
            New Collection
          </Button>
        </div>
      )}

      {collections.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-muted-foreground mb-4">You don't have any collections yet</p>
          {showCreateButton && (
            <Button onClick={() => setShowCreateDialog(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Create Your First Collection
            </Button>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {collections.map((collection) => (
            <CollectionCard key={collection.id} collection={collection} />
          ))}
        </div>
      )}

      {showCreateDialog && (
        <CreateCollectionDialog onClose={() => setShowCreateDialog(false)} />
      )}

      {selectedCollectionId && (
        <div className="fixed inset-0 z-50 flex justify-end">
          <div
            className="absolute inset-0 bg-black/50"
            onClick={() => setSelectedCollectionId(null)}
          />
          <div className="relative w-full max-w-md h-full">
            <CollectionDetailPanel
              collectionId={selectedCollectionId}
              onClose={() => setSelectedCollectionId(null)}
            />
          </div>
        </div>
      )}
    </div>
  )
}

export { CollectionsList }
