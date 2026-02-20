import { useCallback, useEffect, useRef, useState } from 'react'
import { driver, type Config, type Driver } from 'driver.js'
import 'driver.js/dist/driver.css'
import { tours } from '@/lib/tours'

interface UseTourOptions {
  tourId: string
  autoStart?: boolean
  onComplete?: () => void
  onSkip?: () => void
}

export function useTour({ tourId, autoStart = false, onComplete, onSkip }: UseTourOptions) {
  const driverObj = useRef<Driver | null>(null)
  const [isRunning, setIsRunning] = useState(false)
  const [isCompleted, setIsCompleted] = useState(() => {
    const completedTours = localStorage.getItem('fluxbase_completed_tours') || ''
    return completedTours.split(',').includes(tourId)
  })

  const tour = tours[tourId]
  const storageKey = 'fluxbase_completed_tours'

  const startTour = useCallback(() => {
    if (!tour || isRunning) return

    // Small delay to ensure DOM is ready
    setTimeout(() => {
      setIsRunning(true)

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
        onDestroyed: (_element, _step, { driver: drv }) => {
          setIsRunning(false)
          const state = drv.getState()
          const wasCompleted = state.activeIndex === tour.steps.length - 1

          if (wasCompleted) {
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
        },
      } as Config)

      driverObj.current?.drive()
    }, 300)
  }, [tour, isRunning, tourId, storageKey, onComplete, onSkip])

  const skipTour = () => {
    driverObj.current?.destroy()
    setIsRunning(false)
  }

  // Auto-start tour if requested and not already completed
  useEffect(() => {
    if (autoStart && !isCompleted && tour) {
      startTour()
    }
  }, [autoStart, isCompleted, tour, startTour])

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
