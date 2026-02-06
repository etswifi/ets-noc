import { useEffect, useRef } from 'react'

interface AddressAutocompleteProps {
  value: string
  onChange: (value: string) => void
  className?: string
  placeholder?: string
}

export default function AddressAutocomplete({
  value,
  onChange,
  className = '',
  placeholder = 'Enter address...'
}: AddressAutocompleteProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const autocompleteRef = useRef<google.maps.places.Autocomplete | null>(null)

  useEffect(() => {
    // Check if Google Maps is loaded
    if (!window.google || !window.google.maps || !window.google.maps.places) {
      console.warn('Google Maps not loaded')
      return
    }

    if (!inputRef.current) return

    // Initialize autocomplete
    autocompleteRef.current = new window.google.maps.places.Autocomplete(inputRef.current, {
      types: ['address'],
      fields: ['formatted_address', 'address_components']
    })

    // Listen for place selection
    const listener = autocompleteRef.current.addListener('place_changed', () => {
      const place = autocompleteRef.current?.getPlace()
      if (place && place.formatted_address) {
        onChange(place.formatted_address)
      }
    })

    return () => {
      if (listener) {
        window.google.maps.event.removeListener(listener)
      }
    }
  }, [onChange])

  return (
    <input
      ref={inputRef}
      type="text"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className={className}
      placeholder={placeholder}
    />
  )
}
