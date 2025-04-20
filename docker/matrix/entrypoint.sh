#!/bin/sh

# This script runs as root to set up GPIO permissions
# before switching to the non-root user to run the application

# Set permissions for GPIO devices
if [ -e /dev/gpiomem ]; then
  chmod 666 /dev/gpiomem
fi

if [ -d /sys/class/gpio ]; then
  chmod -R 777 /sys/class/gpio
fi

# Execute the application as the non-root user
exec /app/matrix 