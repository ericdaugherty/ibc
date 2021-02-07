# IBC Boiler Logger

The IBC Boiler Logger is a tool to monitor and log the status of your internet connected IBC Boiler.

The IBC Boiler must be internet/intranet connected and be accessible. It is not reccomended to expose the IBC Boiler directly to the internet so this tool is best deployed locally.

## Features

The IBC Boiler Logger writes the current boiler status to a CSV file ever n minutes.

## Usage

Download and compile this tool locally.

Usage:
```
Usage:
  ibclogger [OPTIONS]

Application Options:
  -u, --url=      URL of the Boiler, ex -u "http://192.168.10.2/"
  -f, --file=     File name/path of the output CSV file
  -i, --interval= The number of minutes to wait between log outputs.

Help Options:
  -h, --help      Show this help message
```
