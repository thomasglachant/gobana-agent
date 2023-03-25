# Gobana

Gobana is a simple and fast log monitoring tool.

You can use agent to monitor log files and configure alerting.

## Basic usage

First, you must create a configuration file `config.yml`.

Create your own config file using [agent configuration file template](resources/dist/config/gobana_agent.yaml)

Next, run `gobana-agent` using:

```bash
################
#   Makefile   #
################
make start config=config.yaml

################
# Command line #
################
# Build
make build
# Run
./bin/gobana-agent -config=config.yaml
```

## Test

You can use the provided binary to generate test data.

```bash
# Arguments :
# - log type (symfony|json)
# - log level (CRITICAL|ERROR|WARNING|INFO|DEBUG)
# - number of lines to generate
# - log file
./bin/gen_random_log.sh symfony CRITICAL 10 ~/Downloads/mylogfile.txt
```

## Regex patterns {#regexes}

* Symfony logs regex : `\\[(?P<date>.+)\\] [a-zA-Z0-9_\\-]+.(?P<level>[a-zA-Z0-9]+): (?P<message>.*)`
* Nginx access
  regex : `'(?im)(?P<ipaddress>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) - .* \[(?P<dateandtime>\d{2}\/[a-z]{3}\/\d{4}:\d{2}:\d{2}:\d{2} (\+|\-)\d{4})\] ((\"(?P<method>GET|POST|HEAD|PUT|DELETE|CONNECT|OPTIONS|TRACE|PATCH) )(?P<url>.+)(http\/1\.1")) (?P<statuscode>\d{3}) (?P<bytessent>\d+) (?P<http_referer>[^\s]+)\"\s\"(?P<user_agent>[^\"]+)\"\s\"(?P<forward_for>[^\"]+)\"'`
