import { useState } from 'react'
import { Search, Menu, User } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

export function SearchBar() {
  const [searchTerm, setSearchTerm] = useState('')
  const [isOpen, setIsOpen] = useState(false)
  const [results, setResults] = useState<any[]>([])

  const handleSearch = (value: string) => {
    setSearchTerm(value)
    // Perform search (mock for now)
    const mockResults = [
      { id: '1', type: 'user', name: 'John Doe', email: 'john@example.com' },
      { id: '2', type: 'table', name: 'Products', count: 150 },
      { id: '3', type: 'api', name: 'Get Users', method: 'GET' },
    ]
    setResults(mockResults)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === 'k') {
      e.preventDefault()
      handleSearch(searchTerm)
    }
  }

  return (
    <div className='flex items-center gap-4 flex-1 bg-muted/30 px-4 py-2 border-border relative'>
      <div className='flex w-full max-w-2xl items-center gap-3'>
        {/* Search Input */}
        <div className='flex items-center gap-2 flex-1 w-full max-w-xl'>
          <Search className='w-4 h-4 text-muted-foreground' />
          <Input
            placeholder='Search tables, users, API docs...'
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            onKeyDown={handleKeyDown}
            className='flex-1 bg-background/95 rounded-lg border-border-gray-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-inset-0 focus-visible:transition-colors'
            autoFocus
          />
          <Button
            variant='ghost'
            size='icon'
            onClick={() => setIsOpen(!isOpen)}
            className='h-8 w-8 text-muted-foreground hover:text-foreground'
          >
            <Menu className='h-4 w-4' />
          </Button>
        </div>

        {/* Search Results Dropdown */}
        {results.length > 0 && isOpen && (
          <div className='absolute top-full mt-2 w-56 bg-background rounded-lg shadow-xl border z-40'>
            <div className='max-h-[60vh] overflow-auto'>
              {results.map((result) => (
                <button
                  key={result.id}
                  onClick={() => {
                    setResults([result])
                    setIsOpen(false)
                  }}
                  className='w-full px-4 py-2 hover:bg-muted/50 rounded-lg flex items-center justify-between text-sm text-foreground cursor-pointer'
                >
                  <div className='flex items-center gap-2'>
                    <div
                      className={`p-2 rounded-lg ${
                        result.type === 'user'
                          ? 'bg-blue-100'
                          : result.type === 'table'
                          ? 'bg-green-100'
                          : result.type === 'api'
                          ? 'bg-purple-100'
                          : 'bg-gray-100'
                      }`}
                    >
                      {result.type === 'user' && <User className='h-4 w-4' />}
                      {result.type === 'table' && <span className='text-xs font-mono'>{result.count}</span>}
                      {result.type === 'api' && <span className='text-xs font-mono'>{result.method}</span>}
                    </div>
                    <span className='ml-2 text-muted-foreground'>{result.name}</span>
                  </div>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
