# IBC Boiler Monitor

The IBC Boiler Monitor is a tool to monitor the status of your internet connected IBC Boiler.

The IBC Boiler must be internet/intranet connected and be accessible. It is not reccomended to expose the IBC Boiler directly to the internet so this tool is best deployed locally.

## Features

### Daily Stats
The IBC Monitor Tool will record the numer of cycles for the boiler overall, as well as for each load, daily.

This will be written to a CSV file specified with the -o or --csvOutputFile command line paramter. An example of the output CSV file is:
```
Date,Total,Load 1,Load 2,Load 3,Load 4
2018-04-01,9,2,7,0,0
2018-04-02,10,3,7,0,0
2018-04-03,9,1,8,0,0
```

### Weekly Notification

Coming Soon

### Error and Warning Monitor
If your boiler starts issuing warnings or errors, it is important to be notified quickly. The IBC Monitor tool will check the status of the boiler every 5 minutes and send an email

Download and compile this tool locally or use the [Docker image](https://hub.docker.com/r/ericdaugherty/ibcmonitor).

Usage:
```
Usage:
  ibcmonitor [OPTIONS]

Application Options:
  -u, --url=              URL of the Boiler, ex -u "http://192.168.10.2/"
  -o, --csvOutputFile=    Path to csv of daily cycles.
  -f, --emailFrom=        The email address to use for the FROM setting.
  -t, --emailTo=          The email address to use for the TO setting. Can specify multiple.
      --emailSubject=     The email address to use for the FROM setting. (default: IBC Boiler Alert)
  -s, --emailServer=      The SMTP Server to use to send the email.
      --emailServerPort=  The port to use to connect to the SMTP Server (default: 587)
  -l, --emailUser=        The SMTP Username to use, if needed.
  -p, --emailPass=        The SMTP Password to use, if needed.
  -m, --emailMuteMinutes= The amount of time to wait between sending emails. (default: 60)
```
To run via Docker, first pull the image:
```
docker pull ericdaugherty/ibcmonitor
```

Then start a new container: 
```
docker run --restart always -d -e TZ='America/Denver' -v `pwd`:/root/logs --name ibcmonitor ericdaugherty/ibcmonitor -u <URL of Boiler> -o logs/daily.csv -f <email address> -t <emamil address> -s <smtp.example.com>
```

You can also specify a username and password for the SMTP server with --emailUser and --emailPass
