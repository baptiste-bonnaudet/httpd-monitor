# HTTPD Monitor 

Monitor traffic using an access log, not really useful but a good example of ring buffer, log parsing, time layouts, subroutines and mutexes in go!

Displays stats every 10s about the traffic during those 10s: the sections of the web site with the most hits, as well as interesting summary statistics on the traffic as a whole. 

A user can keep the app running and monitor the log file continuously Whenever total traffic for the past 2 minutes exceeds a certain number on average an alert is displayed. Whenever the total traffic drops again below that value on average for the past 2 minutes, another message detailing that the alert has recovered is displayed. 

## Usage

Configure the app using the `.env`.

Then use the `make` commands to pilot the test env.

```
make

build                          Build the test environment
down                           Start the test environment
help                           Display this help message
logs                           Show and follow the containers logs
up                             Start the test environment
```

## Todo / improvements
- parse commandline args and sanitize inputs
- exclude logs older than our time window during initial parsing.
- write unit tests for AlertAndNotify
- write integration/acceptance tests
