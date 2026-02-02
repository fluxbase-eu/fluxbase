import { createFileRoute } from '@tanstack/react-router'
import { Bot } from 'lucide-react'
import { AIProvidersTab } from '@/components/ai-providers/ai-providers-tab'

const AIProvidersPage = () => {
  return (
    <div className='flex h-full flex-col'>
      {/* Header */}
      <div className='bg-background flex items-center justify-between border-b px-6 py-4'>
        <div className='flex items-center gap-3'>
          <div className='bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg'>
            <Bot className='text-primary h-5 w-5' />
          </div>
          <div>
            <h1 className='text-xl font-semibold'>AI Providers</h1>
            <p className='text-muted-foreground text-sm'>
              Configure AI providers for chatbots and intelligent features
            </p>
          </div>
        </div>
      </div>

      <div className='flex-1 overflow-auto p-6'>
        <AIProvidersTab />
      </div>
    </div>
  )
}

export const Route = createFileRoute('/_authenticated/ai-providers/')({
  component: AIProvidersPage,
})
