# IBC Boiler API

The ibc package provices a go interface to an IBC Boiler.

This package simplifies access to IBC Boiler status information. At this time, the interface is read-only.

The IBC Boiler must be internet/intranet connected and be accessible. It is not reccomended to expose the IBC Boiler to the internet so this library is best accessed via intranet.

This repository also provides the ibcmonitor tool, which allows you to monitor the statistics and status of your boiler, and can be deployed as a docker image. Check it out [here](https://github.com/ericdaugherty/ibc/tree/master/tools/cmd/ibcmonitor)

See the cmd/ibcstatus tool for sample usage.