interface LogoProps {
  className?: string
  size?: 'sm' | 'md' | 'lg'
  showText?: boolean
}

export default function Logo({ className = '', size = 'md', showText = true }: LogoProps) {
  const sizes = {
    sm: 'h-8',
    md: 'h-12',
    lg: 'h-16',
  }

  return (
    <div className={`${sizes[size]} ${className} flex items-center gap-3`}>
      <img
        src="/ets-logo.png"
        alt="ETS Logo"
        className={`${sizes[size]} object-contain`}
      />
      {showText && (
        <span className={`font-bold ${size === 'lg' ? 'text-2xl' : size === 'md' ? 'text-xl' : 'text-lg'} text-gray-900 dark:text-white`}>
          NOC
        </span>
      )}
    </div>
  )
}
