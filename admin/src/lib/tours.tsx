export interface TourStep {
  element: string | (() => Element | null)
  popover?: {
    className?: string
    side?: 'top' | 'right' | 'bottom' | 'left'
    align?: 'start' | 'center' | 'end'
  }
  title: string
  description: string | React.ReactNode
}

export interface Tour {
  id: string
  name: string
  steps: TourStep[]
}

export const tours: Record<string, Tour> = {
  dashboard: {
    id: 'dashboard',
    name: 'Dashboard Overview',
    steps: [
      {
        element: '.health-status-card',
        title: 'Health Status',
        description: 'Monitor all your backend services in real-time. Green means everything is operational.',
      },
      {
        element: '.metrics-cards',
        title: 'Key Metrics',
        description: 'Track requests per minute, error rates, response times, and active users at a glance.',
      },
      {
        element: '.activity-feed',
        title: 'Activity Feed',
        description: 'See recent events including user signups, function executions, and system alerts.',
      },
      {
        element: '.quick-actions',
        title: 'Quick Actions',
        description: 'Access common tasks quickly. Actions adapt based on your current setup state.',
      },
    ],
  },

  tables: {
    id: 'tables',
    name: 'Database Tables',
    steps: [
      {
        element: '[data-testid="table-selector"]',
        title: 'Table Selector',
        description: 'Browse all your database tables organized by schema. Click to view table data.',
      },
      {
        element: '[data-testid="create-table-btn"]',
        title: 'Create Tables',
        description: 'Define new database tables with columns, types, and constraints.',
      },
      {
        element: '[data-testid="table-viewer"]',
        title: 'Table Data',
        description: 'View, filter, and edit your table data. Supports pagination and bulk operations.',
      },
    ],
  },

  functions: {
    id: 'functions',
    name: 'Edge Functions',
    steps: [
      {
        element: '[data-testid="functions-list"]',
        title: 'Functions List',
        description: 'View all deployed edge functions with their status and execution history.',
      },
      {
        element: '[data-testid="create-function-btn"]',
        title: 'Create Functions',
        description: 'Deploy new TypeScript/JavaScript functions with the Deno runtime.',
      },
      {
        element: '[data-testid="function-executions"]',
        title: 'Execution Logs',
        description: 'Monitor function executions, view logs, and debug errors.',
      },
    ],
  },

  users: {
    id: 'users',
    name: 'User Management',
    steps: [
      {
        element: '[data-testid="users-table"]',
        title: 'Users Table',
        description: 'View all application users with their verification status and last sign-in time.',
      },
      {
        element: '[data-testid="invite-user-btn"]',
        title: 'Invite Users',
        description: 'Send invitations to new users. They\'ll receive an email to set up their account.',
      },
      {
        element: '[data-testid="user-tabs"]',
        title: 'User Types',
        description: 'Switch between application users and dashboard administrators.',
      },
    ],
  },

  policies: {
    id: 'policies',
    name: 'RLS Policies',
    steps: [
      {
        element: '[data-testid="policies-table"]',
        title: 'Policy Table',
        description: 'View and manage Row Level Security policies that control data access.',
      },
      {
        element: '[data-testid="create-policy-btn"]',
        title: 'Create Policies',
        description: 'Define policies using SQL to restrict data access based on user attributes.',
      },
      {
        element: '[data-testid="policy-editor"]',
        title: 'Policy Editor',
        description: 'Write and test policies with syntax highlighting and validation.',
      },
    ],
  },

  storage: {
    id: 'storage',
    name: 'File Storage',
    steps: [
      {
        element: '[data-testid="buckets-list"]',
        title: 'Storage Buckets',
        description: 'Organize files into buckets for different purposes (uploads, avatars, documents, etc.).',
      },
      {
        element: '[data-testid="create-bucket-btn"]',
        title: 'Create Buckets',
        description: 'Create new storage buckets with access policies and size limits.',
      },
      {
        element: '[data-testid="files-browser"]',
        title: 'File Browser',
        description: 'Upload, download, and manage files with drag-and-drop support.',
      },
    ],
  },
}
