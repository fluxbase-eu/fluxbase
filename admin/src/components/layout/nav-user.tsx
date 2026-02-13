import { LogOut, User, LifeBuoy } from 'lucide-react'
import { useNavigate } from '@tanstack/react-router'
import { logout } from '@/lib/auth'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { SidebarMenu, SidebarMenuItem } from '@/components/ui/sidebar'
import { Separator } from '@/components/ui/separator'

// Component to display user name in sidebar footer
type NavUserProps = {
  user: {
    name: string
    email: string
    avatar: string
  }
}

export function NavUser({ user }: NavUserProps) {
  const navigate = useNavigate()

  const showTour = () => {
    // Determine which tour to show based on current path
    const path = window.location.pathname
    let tourId = 'dashboard'

    if (path.startsWith('/tables')) tourId = 'tables'
    else if (path.startsWith('/functions')) tourId = 'functions'
    else if (path.startsWith('/users')) tourId = 'users'
    else if (path.startsWith('/policies')) tourId = 'policies'
    else if (path.startsWith('/storage')) tourId = 'storage'

    // Navigate to tour page with tourId parameter
    navigate({ to: `/tour?t=${tourId}` })
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <div className='flex items-center gap-2 px-2 py-1.5'>
          <Avatar className='h-8 w-8'>
            <AvatarImage src={user.avatar} alt={user.name} />
            <AvatarFallback>
              <User className='h-4 w-4' />
            </AvatarFallback>
          </Avatar>
          <div className='grid flex-1 text-start text-sm leading-tight'>
            <span className='truncate font-semibold'>{user.name}</span>
            <span className='text-muted-foreground truncate text-xs'>
              {user.email}
            </span>
          </div>
          <Button
            variant='ghost'
            size='icon'
            className='h-8 w-8'
            onClick={logout}
            title='Sign out'
          >
            <LogOut className='h-4 w-4' />
          </Button>
        </div>
      </SidebarMenuItem>
      <SidebarMenuItem>
        <Separator className='my-1' />
        <button
          onClick={showTour}
          className='flex w-full items-center gap-2 rounded-md px-2 py-2 text-sm transition-colors hover:bg-accent'
        >
          <LifeBuoy className='h-4 w-4 text-muted-foreground' />
          <span>Show Tour</span>
        </button>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
