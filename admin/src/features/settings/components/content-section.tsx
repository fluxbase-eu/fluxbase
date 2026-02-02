import type { LucideIcon } from 'lucide-react'

type ContentSectionProps = {
  title: string
  desc: string
  children: React.JSX.Element
  icon?: LucideIcon
}

export function ContentSection({
  title,
  desc,
  children,
  icon: Icon,
}: ContentSectionProps) {
  return (
    <div className='flex h-full flex-col'>
      <div className='bg-background flex items-center justify-between border-b px-6 py-4'>
        <div className='flex items-center gap-3'>
          {Icon && (
            <div className='bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg'>
              <Icon className='text-primary h-5 w-5' />
            </div>
          )}
          <div>
            <h1 className='text-xl font-semibold'>{title}</h1>
            <p className='text-muted-foreground text-sm'>{desc}</p>
          </div>
        </div>
      </div>
      <div className='flex-1 overflow-auto p-6'>
        <div className='lg:max-w-xl'>{children}</div>
      </div>
    </div>
  )
}
