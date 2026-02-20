import * as React from 'react'
import * as ProgressPrimitive from '@radix-ui/react-progress'
import { cva, type VariantProps } from 'class-variance-authority'

import { cn } from '@/lib/utils'

const progressVariants = cva({
  base: {
    className: 'relative w-full overflow-hidden rounded-full bg-primary/20',
  },
})

const Progress = React.forwardRef<
  React.ElementRef<typeof ProgressPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof ProgressPrimitive.Root> & VariantProps<typeof progressVariants>
>(({ className, value = 0, ...props }, ref) => (
  <ProgressPrimitive.Root
    ref={ref}
    className={cn(progressVariants({ className }), 'h-2 w-full')}
    value={value}
    {...props}
  >
    <ProgressPrimitive.Indicator className='h-full w-full flex-1 bg-primary transition-all' style={{ transform: `translateX(-${100 - (value || 0)}%)` }} />
  </ProgressPrimitive.Root>
))
Progress.displayName = ProgressPrimitive.Root.displayName

export { Progress }
