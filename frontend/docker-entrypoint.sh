#!/bin/sh
set -e

# Replace the placeholder with the actual API key in index.html
if [ -n "$GOOGLE_MAPS_API_KEY" ]; then
  sed -i "s/__GOOGLE_MAPS_API_KEY__/$GOOGLE_MAPS_API_KEY/g" /usr/share/nginx/html/index.html
  echo "Google Maps API key injected"
else
  echo "No Google Maps API key provided, autocomplete will not work"
fi

# Start nginx
exec nginx -g 'daemon off;'
