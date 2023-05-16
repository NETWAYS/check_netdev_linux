**Warning:** This is a prototype and not finished and not very well tested.

check_netdev_linux
==================

Check Linux network interfaces (mostly to read performance data)

# Basics
This is a small plugin to check the network devices on a linux machine. It can check whether the devices are online and will probably include a way to define threshholds for
the amount of traffic in the future.
Currently it mostly gets the statistics, so one can paint pretty graphs.

# Building
```
go build
```

# Usage
```
# Read the statistics for all available network interfaces and say it's ok
./check_netdev_linux

# Check whether all network interfaces are online
./check_netdev_linux -c

# Get diffs for network statistics for the time frame of 5 seconds
./check_netdev_linux -m 5

# Just get statistics for ethernet interfaces (all devices matching the expression)
./check_netdev_linux -i 'en.'

# Exclude the first devices (which end on a 0)
./check_netdev_linux -e '.0'

# Only read from openvpn interface (a device called openvpn to be specific)
./check_netdev_linux -I openvpn
```
