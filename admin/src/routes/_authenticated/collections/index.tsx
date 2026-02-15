import { createFileRoute } from '@tanstack/react-router'
import { CollectionsList } from '@/features/collections'

export const Route = createFileRoute('/_authenticated/collections/')({
  component: CollectionsPage,
})

function CollectionsPage() {
  return (
    <div className="container mx-auto py-6">
      <div className="mb-6">
        <h1 className="text-3xl font-bold tracking-tight">Collections</h1>
        <p className="text-muted-foreground mt-2">
          Organize your knowledge bases into collections to keep them organized and easy to find.
        </p>
      </div>

      <CollectionsList showCreateButton={true} />
    </div>
  )
}
