import {
  LayoutDashboard,
  Database,
  GitFork,
  Code,
  Code2,
  ScrollText,
  Activity,
  Users,
  Shield,
  Zap,
  FileCode,
  FolderOpen,
  Radio,
  ListTodo,
  Terminal,
  Bot,
  BookOpen,
  Wrench,
  Key,
  KeyRound,
  ShieldAlert,
  ShieldCheck,
  Webhook,
  Lock,
  Settings,
  Palette,
  Puzzle,
  Mail,
  HardDrive,
  Command,
} from 'lucide-react'
import { type SidebarData } from '../types'

export const sidebarData: SidebarData = {
  user: {
    name: 'Admin',
    email: 'admin@fluxbase.eu',
    avatar: '',
  },
  teams: [
    {
      name: 'Fluxbase',
      logo: Command,
      plan: 'Backend as a Service',
    },
  ],
  navGroups: [
    {
      title: 'Overview',
      items: [
        {
          title: 'Dashboard',
          url: '/',
          icon: LayoutDashboard,
        },
      ],
    },
    {
      title: 'Database',
      items: [
        {
          title: 'Tables',
          url: '/tables',
          icon: Database,
        },
        {
          title: 'Schema Viewer',
          url: '/schema',
          icon: GitFork,
        },
        {
          title: 'SQL Editor',
          url: '/sql-editor',
          icon: Code,
        },
      ],
    },
    {
      title: 'Monitoring',
      items: [
        {
          title: 'Log Stream',
          url: '/logs',
          icon: ScrollText,
        },
        {
          title: 'Monitoring',
          url: '/monitoring',
          icon: Activity,
        },
      ],
    },
    {
      title: 'Users & Authentication',
      items: [
        {
          title: 'Users',
          url: '/users',
          icon: Users,
        },
        {
          title: 'Authentication',
          url: '/authentication',
          icon: Shield,
        },
      ],
    },
    {
      title: 'API & Services',
      items: [
        {
          title: 'API Explorer',
          url: '/api/rest',
          icon: Code2,
        },
        {
          title: 'Realtime',
          url: '/realtime',
          icon: Radio,
        },
        {
          title: 'Storage',
          url: '/storage',
          icon: FolderOpen,
        },
        {
          title: 'Functions',
          url: '/functions',
          icon: FileCode,
        },
        {
          title: 'Jobs',
          url: '/jobs',
          icon: ListTodo,
        },
        {
          title: 'RPC',
          url: '/rpc',
          icon: Terminal,
        },
        {
          title: 'AI Chatbots',
          url: '/chatbots',
          icon: Bot,
        },
        {
          title: 'Knowledge Bases',
          url: '/knowledge-bases',
          icon: BookOpen,
        },
        {
          title: 'MCP Tools',
          url: '/mcp-tools',
          icon: Wrench,
        },
      ],
    },
    {
      title: 'Security',
      items: [
        {
          title: 'RLS Policies',
          url: '/policies',
          icon: ShieldAlert,
        },
        {
          title: 'Security Settings',
          url: '/security-settings',
          icon: ShieldCheck,
        },
        {
          title: 'Secrets',
          url: '/secrets',
          icon: Lock,
        },
        {
          title: 'Client Keys',
          url: '/client-keys',
          icon: Key,
        },
        {
          title: 'Service Keys',
          url: '/service-keys',
          icon: KeyRound,
        },
        {
          title: 'Webhooks',
          url: '/webhooks',
          icon: Webhook,
        },
      ],
    },
    {
      title: 'Configuration',
      items: [
        {
          title: 'Features',
          url: '/features',
          icon: Zap,
        },
        {
          title: 'Extensions',
          url: '/extensions',
          icon: Puzzle,
        },
        {
          title: 'Email',
          url: '/email-settings',
          icon: Mail,
        },
        {
          title: 'Storage',
          url: '/storage-config',
          icon: HardDrive,
        },
        {
          title: 'AI Providers',
          url: '/ai-providers',
          icon: Bot,
        },
        {
          title: 'Database',
          url: '/database-config',
          icon: Database,
        },
      ],
    },
    {
      title: 'Account settings',
      items: [
        {
          title: 'Account',
          url: '/settings',
          icon: Settings,
        },
        {
          title: 'Appearance',
          url: '/settings/appearance',
          icon: Palette,
        },
      ],
    },
  ],
}
