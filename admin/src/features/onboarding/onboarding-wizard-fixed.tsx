import { useState, useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  Sparkles,
  Check,
  Users,
  Database,
  FolderOpen,
  FileCode,
  Shield,
  ArrowRight,
  ArrowLeft,
} from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Main } from '@/components/layout/main'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

type Step = 'welcome' | 'features' | 'api-keys'

const ONBOARDING_STORAGE_KEY = 'fluxbase_onboarding_completed'
const ONBOARDING_STATE_KEY = 'fluxbase_onboarding_state'

interface OnboardingState {
  projectName: string
  environment: 'development' | 'staging' | 'production'
  features: {
    auth: boolean
    database: boolean
    storage: boolean
    functions: boolean
  }
}

const defaultState: OnboardingState = {
  projectName: '',
  environment: 'development',
  features: {
    auth: true,
    database: true,
    storage: false,
    functions: false,
  },
}

export function OnboardingWizard() {
  const navigate = useNavigate()

  const [currentStep, setCurrentStep] = useState<Step>('welcome')
  const [state, setState] = useState<OnboardingState>(defaultState)

  // Check if onboarding was already completed
  useEffect(() => {
    const completed = localStorage.getItem(ONBOARDING_STORAGE_KEY)
    if (completed === 'true') {
      navigate({ to: '/' })
      return
    }

    // Load saved state
    const savedState = localStorage.getItem(ONBOARDING_STATE_KEY)
    if (savedState) {
      try {
        setState(JSON.parse(savedState))
      } catch {
        // Use default state
      }
    }
  }, [navigate])

  const steps: { id: Step; title: string; description: string }[] = [
    {
      id: 'welcome',
      title: 'Welcome to Fluxbase',
      description: 'Let\'s get your backend ready in 3 simple steps',
    },
    {
      id: 'features',
      title: 'Enable Key Features',
      description: 'Choose the services you need for your application',
    },
    {
      id: 'api-keys',
      title: 'Get Your API Keys',
      description: 'Your credentials for connecting your application',
    },
  ]

  const currentStepIndex = steps.findIndex((s) => s.id === currentStep)

  const saveState = () => {
    localStorage.setItem(ONBOARDING_STATE_KEY, JSON.stringify(state))
  }

  const updateState = (updates: Partial<OnboardingState>) => {
    setState((prev) => {
      const newState = { ...prev, ...updates }
      saveState()
      return newState
    })
  }

  const handleNext = () => {
    if (currentStep === 'welcome') {
      if (!state.projectName.trim()) {
        toast.error('Please enter a project name')
        return
      }
      setCurrentStep('features')
    } else if (currentStep === 'features') {
      setCurrentStep('api-keys')
    } else if (currentStep === 'api-keys') {
      // Complete onboarding
      localStorage.setItem(ONBOARDING_STORAGE_KEY, 'true')
      toast.success('Setup complete! Welcome to Fluxbase')
      navigate({ to: '/' })
    }
  }

  const handleBack = () => {
    if (currentStep === 'features') {
      setCurrentStep('welcome')
    } else if (currentStep === 'api-keys') {
      setCurrentStep('features')
    }
  }

  const handleSkip = () => {
    localStorage.setItem(ONBOARDING_STORAGE_KEY, 'true')
    navigate({ to: '/' })
  }

  return (
    <Main className='flex items-center justify-center p-6'>
      <Card className='w-full max-w-2xl'>
        <CardContent className='p-8'>
          {/* Progress */}
          <div className='mb-8'>
            <div className='mb-2 flex items-center justify-between'>
              <div className='flex items-center gap-2'>
                <Sparkles className='h-5 w-5 text-primary' />
                <h2 className='text-lg font-semibold'>{steps[currentStepIndex].title}</h2>
              </div>
              <button
                onClick={handleSkip}
                className='text-muted-foreground hover:text-foreground text-sm'
              >
                Skip setup
              </button>
            </div>
            <p className='text-muted-foreground text-sm'>
              {steps[currentStepIndex].description}
            </p>
            <div className='mt-4 flex items-center gap-2'>
              {steps.map((step, index) => (
                <div
                  key={step.id}
                  className={cn(
                    'flex-1 h-1 rounded-full',
                    index <= currentStepIndex ? 'bg-primary' : 'bg-muted',
                    index < currentStepIndex && 'bg-primary/50'
                  )}
                />
              ))}
            </div>
            <div className='mt-2 flex justify-between text-xs'>
              {steps.map((step, index) => (
                <span
                  key={step.id}
                  className={cn(
                    'font-medium',
                    index === currentStepIndex
                      ? 'text-primary'
                      : 'text-muted-foreground'
                  )}
                >
                  {index + 1}
                </span>
              ))}
            </div>
          </div>

          {/* Step Content */}
          {currentStep === 'welcome' && (
            <div className='space-y-6'>
              <div>
                <Label htmlFor='projectName'>Project Name</Label>
                <Input
                  id='projectName'
                  placeholder='e.g., my-awesome-app'
                  value={state.projectName}
                  onChange={(e) => updateState({ projectName: e.target.value })}
                  autoFocus
                  className='h-11'
                />
                <p className='text-muted-foreground mt-1.5 text-xs'>
                  A name to identify your project
                </p>
              </div>

              <div>
                <Label>Environment</Label>
                <div className='mt-2 grid grid-cols-3 gap-3'>
                  {(
                    [
                      { value: 'development', label: 'Development' },
                      { value: 'staging', label: 'Staging' },
                      { value: 'production', label: 'Production' },
                    ] as const
                  ).map((env) => (
                    <button
                      key={env.value}
                      onClick={() => updateState({ environment: env.value })}
                      className={cn(
                        'flex flex-col items-center gap-2 rounded-lg border-2 p-4 transition-colors',
                        state.environment === env.value
                          ? 'border-primary bg-primary/5'
                          : 'border-border hover:border-primary/50'
                      )}
                    >
                      {state.environment === env.value && (
                        <div className='bg-primary text-primary-foreground flex h-5 w-5 items-center justify-center rounded-full'>
                          <Check className='h-3 w-3' />
                        </div>
                      )}
                      <span
                        className={cn(
                          'font-medium',
                          state.environment === env.value
                            ? 'text-primary'
                            : 'text-muted-foreground'
                        )}
                      >
                        {env.label}
                      </span>
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}

          {currentStep === 'features' && (
            <div className='space-y-4'>
              <p className='text-muted-foreground text-sm'>
                Select the features you want to enable. You can always enable more later.
              </p>

              <div className='space-y-3'>
                {[
                  {
                    key: 'auth',
                    icon: <Shield className='h-5 w-5' />,
                    title: 'User Authentication',
                    description: 'JWT-based auth with OAuth, SAML, and magic links',
                    default: true,
                  },
                  {
                    key: 'database',
                    icon: <Database className='h-5 w-5' />,
                    title: 'PostgreSQL Database',
                    description: 'Full SQL database with Row Level Security',
                    default: true,
                  },
                  {
                    key: 'storage',
                    icon: <FolderOpen className='h-5 w-5' />,
                    title: 'File Storage',
                    description: 'S3-compatible storage with bucket management',
                    default: false,
                  },
                  {
                    key: 'functions',
                    icon: <FileCode className='h-5 w-5' />,
                    title: 'Edge Functions',
                    description: 'Deno runtime for TypeScript/JavaScript functions',
                    default: false,
                  },
                ].map((feature) => (
                  <button
                    key={feature.key}
                    onClick={() =>
                      updateState({
                        features: {
                          ...state.features,
                          [feature.key]: !state.features[
                            feature.key as keyof typeof state.features
                          ],
                        },
                      })
                    }
                    className={cn(
                      'flex items-start gap-4 rounded-lg border-2 p-4 text-left transition-colors',
                      state.features[feature.key as keyof typeof state.features]
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    )
                    }
                  >
                    <div className='flex-1'>
                      <div className='flex items-center gap-3'>
                        <div
                          className={cn(
                            'flex items-center justify-center rounded-lg p-2',
                            state.features[
                              feature.key as keyof typeof state.features
                            ]
                              ? 'bg-primary text-primary-foreground'
                              : 'bg-muted'
                          )}
                        >
                          {feature.icon}
                        </div>
                        <div className='text-left'>
                          <div className='font-medium'>{feature.title}</div>
                          <div className='text-muted-foreground text-xs'>
                            {feature.description}
                          </div>
                        </div>
                      </div>
                    </div>
                    <div className='flex-1 justify-end'>
                      <Checkbox
                        checked={
                          state.features[
                            feature.key as keyof typeof state.features
                          ]
                        }
                        onCheckedChange={(checked) =>
                          updateState({
                            features: {
                              ...state.features,
                              [feature.key]: checked === true,
                            },
                          })
                        }
                        className='h-5 w-5'
                      />
                    </div>
                  </button>
                ))}
              </div>
            </div>
          )}

          {currentStep === 'api-keys' && (
            <div className='space-y-6'>
              <div className='bg-muted/30 rounded-lg border p-6'>
                <p className='mb-4 text-sm font-medium'>
                  Your API credentials are ready! Use these to connect your
                  application.
                </p>

                <div className='space-y-4'>
                  <div>
                    <Label>Project URL</Label>
                    <div className='mt-1.5 flex items-center gap-2'>
                      <code className='bg-background flex-1 rounded border px-3 py-2 text-sm'>
                        {window.location.origin}
                      </code>
                      <Button
                        variant='outline'
                        size='sm'
                        onClick={() => {
                          navigator.clipboard.writeText(window.location.origin)
                          toast.success('URL copied to clipboard')
                        }}
                      >
                        Copy
                      </Button>
                    </div>
                  </div>

                  <div>
                    <Label>Anonymous Key (Public)</Label>
                    <div className='mt-1.5 flex items-center gap-2'>
                      <code className='bg-background flex-1 break-all rounded border px-3 py-2 text-sm font-mono'>
                        {/* TODO: Get actual anon key from API */}
                        eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
                      </code>
                      <Button
                        variant='outline'
                        size='sm'
                        onClick={() => {
                          navigator.clipboard.writeText(
                            'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
                          )
                          toast.success('Key copied to clipboard')
                        }}
                      >
                        Copy
                      </Button>
                    </div>
                  </div>
                </div>

                <div className='bg-yellow-500/10 border-yellow-500/20 rounded-lg border p-4'>
                  <p className='text-muted-foreground text-xs'>
                    <Shield className='text-yellow-600 dark:text-yellow-400 mr-1.5 inline h-3 w-3' />
                    <strong>Security Note:</strong> The anonymous key provides
                    public access to your database. Configure Row Level Security (RLS)
                    policies to restrict access based on user authentication.
                  </p>
                </div>
              </div>

              <div>
                <h3 className='mb-3 text-sm font-medium'>Next Steps</h3>
                <div className='space-y-2 text-sm'>
                  <a
                    href='/docs/guides/database'
                    target='_blank'
                    rel='noopener noreferrer'
                    className='flex items-center gap-2 text-primary hover:underline'
                  >
                    <Database className='h-4 w-4' />
                    <span>Create your first database table</span>
                  </a>
                  <a
                    href='/docs/guides/auth'
                    target='_blank'
                    rel='noopener noreferrer'
                    className='flex items-center gap-2 text-primary hover:underline'
                  >
                    <Users className='h-4 w-4' />
                    <span>Set up user authentication</span>
                  </a>
                  <a
                    href='/docs/guides/functions'
                    target='_blank'
                    rel='noopener noreferrer'
                    className='flex items-center gap-2 text-primary hover:underline'
                  >
                    <FileCode className='h-4 w-4' />
                    <span>Deploy your first edge function</span>
                  </a>
                </div>
              </div>
            </div>
          )}

          {/* Navigation Buttons */}
          <div className='mt-8 flex justify-between border-t pt-6'>
            {currentStep !== 'welcome' && (
              <Button variant='outline' onClick={handleBack}>
                <ArrowLeft className='mr-2 h-4 w-4' />
                Back
              </Button>
            )}
            <div className='flex-1' />
            <Button onClick={handleNext} className='gap-2'>
              {currentStep === 'api-keys' ? (
                <>
                  <Check className='h-4 w-4' />
                  Go to Dashboard
                </>
              ) : (
                <>
                  Continue
                  <ArrowRight className='h-4 w-4' />
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>
    </Main>
  )
}
