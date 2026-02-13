import { useEffect, useRef, useState } from 'react'
import { driver } from 'driver.js'
import 'driver.js/dist/driver.css'
import type { Config } from 'driver.js'
import { tours } from '@/lib/tours'

interface UseTourOptions {
  tourId: string
  autoStart?: boolean
  onComplete?: () => void
  onSkip?: () => void
}

export function useTour({ tourId, autoStart = false, onComplete, onSkip }: UseTourOptions) {
  const driverObj = useRef<any>(null)
  const [isRunning, setIsRunning] = useState(false)
  const [isCompleted, setIsCompleted] = useState(() => {
    const completedTours = localStorage.getItem('fluxbase_completed_tours') || ''
    return completedTours.split(',').includes(tourId)
  })

  const tour = tours[tourId]
  const storageKey = 'fluxbase_completed_tours'

  const startTour = () => {
    if (!tour || isRunning) return

    setIsRunning(true)

    // Small delay to ensure DOM is ready
    setTimeout(() => {
      driverObj.current = driver({
        showProgress: true,
        steps: tour.steps,
        nextBtnText: 'Next →',
        prevBtnText: '← Back',
        doneBtnText: 'Got it!',
        smoothScroll: true,
        animate: true,
        overlayClickNext: true,
        disableActiveInteraction: false,
        classPrefix: 'driverjs',
        style: `
          .driverjs-popover {
            border-radius: 8px;
            max-width: 320px;
          }
          .driverjs-popover-title {
            font-size: 16px;
            font-weight: 600;
          }
          .driverjs-popover-description {
            font-size: 14px;
            line-height: 1.5;
          }
          .driverjs-next-btn {
            background: hsl(var(--primary));
            color: hsl(var(--primary-foreground));
            border-radius: 6px;
            padding: 8px 16px;
            font-weight: 500;
          }
          .driverjs-prev-btn {
            color: hsl(var(--muted-foreground));
            font-weight: 500;
          }
        `,
      } as Config)

      driverObj.current.drive()

      driverObj.current.on('destroyed', (cancelled: boolean) => {
        setIsRunning(false)

        if (!cancelled) {
          // Tour was completed
          const completedTours = localStorage.getItem(storageKey) || ''
          const newCompleted = [...completedTours.split(','), tourId].filter(Boolean).join(',')
          localStorage.setItem(storageKey, newCompleted)
          setIsCompleted(true)
          onComplete?.()
        } else {
          // Tour was skipped
          onSkip?.()
        }
      })
    }, 300)
  }

  const skipTour = () => {
    driverObj.current?.destroy()
    setIsRunning(false)
  }

  // Auto-start tour if requested and not already completed
  useEffect(() => {
    if (autoStart && !isCompleted && tour) {
      startTour()
    }
  }, [autoStart, isCompleted, tourId])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      driverObj.current?.destroy()
    }
  }, [])

  return {
    isRunning,
    isCompleted,
    startTour,
    skipTour,
    tour,
  }
}
