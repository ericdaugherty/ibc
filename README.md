# IBC Boiler API

The ibc package provices a go interface to an IBC Boiler.

This package simplifies access to IBC Boiler status information. At this time, the interface is read-only.

The IBC Boiler must be internet/intranet connected and be accessible. It is not reccomended to expose the IBC Boiler to the internet so this library is best accessed via intranet.

See the cmd/ibcstatus tool for sample usage.