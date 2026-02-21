import { useState, useEffect, useCallback } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  RefreshCw,
  Database,
  Plus,
  Loader2,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  knowledgeBasesApi,
  type TableSummary,
} from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area'
import { KnowledgeBaseTabs } from '@/components/knowledge-bases/knowledge-base-tabs'

export const Route = createFileRoute('/_authenticated/knowledge-bases/$id/tables')({
  component: KnowledgeBaseTablesPage,
})

function KnowledgeBaseTablesPage() {
  const { id } = Route.useParams()
  const navigate = useNavigate()
  const [tables, setTables] = useState<TableSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [exporting, setExporting] = useState<string | null>(null)
  const [schemaFilter, setSchemaFilter] = useState<string>('all')

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const schema = schemaFilter === 'all' ? undefined : schemaFilter
      const data = await knowledgeBasesApi.listTables(schema)
      setTables(data)
    } catch {
      toast.error('Failed to fetch tables')
    } finally {
      setLoading(false)
    }
  }, [schemaFilter])

  const handleExportTable = async (table: TableSummary) => {
    setExporting(`${table.schema}.${table.name}`)
    try {
      await knowledgeBasesApi.exportTable(id, {
        schema: table.schema,
        table: table.name,
        include_foreign_keys: true,
        include_indexes: true,
        include_sample_rows: false,
      })
      toast.success(`Table ${table.schema}.${table.name} exported successfully`)
    } catch {
      toast.error(`Failed to export table ${table.schema}.${table.name}`)
    } finally {
      setExporting(null)
    }
  }

  useEffect(() => {
    fetchData()
  }, [fetchData])

  // Get unique schemas for filter
  const schemas = ['all', ...Array.from(new Set(tables.map((t) => t.schema))).sort()]

  return (
    <div className="flex flex-1 flex-col gap-6 p-6">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => navigate({ to: '/knowledge-bases/$id', params: { id } })}
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Knowledge Base
        </Button>
      </div>

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Database Tables</h1>
          <p className="text-muted-foreground">
            Export database tables as documents to this knowledge base
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={schemaFilter} onValueChange={setSchemaFilter}>
            <SelectTrigger className="w-[200px]">
              <SelectValue placeholder="Filter by schema" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Schemas</SelectItem>
              {schemas
                .filter((s) => s !== 'all')
                .map((schema) => (
                  <SelectItem key={schema} value={schema}>
                    {schema}
                  </SelectItem>
                ))}
            </SelectContent>
          </Select>
          <Button onClick={fetchData} variant="outline" size="sm">
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Tab Navigation */}
      <KnowledgeBaseTabs activeTab='tables' knowledgeBaseId={id} />

      <Card>
        <CardHeader>
          <CardTitle>Exportable Tables</CardTitle>
          <CardDescription>
            Database tables that can be exported as documents with embedded schema
            information
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex h-96 items-center justify-center">
              <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
            </div>
          ) : tables.length === 0 ? (
            <div className="py-12 text-center">
              <Database className="text-muted-foreground mx-auto mb-4 h-12 w-12" />
              <p className="mb-2 text-lg font-medium">No tables found</p>
              <p className="text-muted-foreground text-sm">
                No tables available for export in the selected schema
              </p>
            </div>
          ) : (
            <ScrollArea className="h-[600px]">
              <div className="min-w-[600px]">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Schema</TableHead>
                      <TableHead>Table Name</TableHead>
                      <TableHead>Columns</TableHead>
                      <TableHead>Foreign Keys</TableHead>
                      <TableHead className="w-[120px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {tables.map((table) => (
                      <TableRow key={`${table.schema}.${table.name}`}>
                        <TableCell className="font-medium">
                          <Badge variant="outline">{table.schema}</Badge>
                        </TableCell>
                        <TableCell>
                          <code className="text-sm">{table.name}</code>
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary">{table.columns}</Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary">{table.foreign_keys}</Badge>
                        </TableCell>
                        <TableCell>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleExportTable(table)}
                            disabled={
                              exporting === `${table.schema}.${table.name}`
                            }
                          >
                            {exporting === `${table.schema}.${table.name}` ? (
                              <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                Exporting
                              </>
                            ) : (
                              <>
                                <Plus className="mr-2 h-4 w-4" />
                                Export
                              </>
                            )}
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
              <ScrollBar orientation="horizontal" />
            </ScrollArea>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
