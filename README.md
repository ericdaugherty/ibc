# IBC Boiler API

The ibc package provices a go interface to an IBC Boiler.

This package simplifies access to IBC Boiler status information. At this time, the interface is read-only.

The IBC Boiler must be internet/intranet connected and be accessible. It is not reccomended to expose the IBC Boiler to the internet so this library is best accessed via intranet.

This repository also provides a set of command line tools that provide basic monitoring and logging functionality.
- [IBC Logger](https://github.com/ericdaugherty/ibc/tree/master/tools/cmd/ibclogger)
- [IBC Monitor](https://github.com/ericdaugherty/ibc/tree/master/tools/cmd/ibcmonitor)
- [IBC Status](https://github.com/ericdaugherty/ibc/tree/master/tools/cmd/ibcstatus)
