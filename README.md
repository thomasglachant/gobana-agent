# Spooter

Spooter is a simple, fast, and powerful log monitoring tool.

You can use agent only to monitor log files and configure alerting,
or you can use agent + server to centralize log monitoring.

## Agent mode

First, you must create a configuration file `config_agent.yml`.

```yaml
#
# Agent configuration
#
debug: false

#
# Metadata about the agent 
application: "MyApp"  # Name of your application (required)
server: "localhost"   # Name of your server (optional)

#
# Parsers are used to read and normalize logs 
# Normalize logs permit to have a uniform format, usable to alerting and analyses.
# 
# There is two kinds of parsers : 
# - `regex` : parse a log line using a regex to capture fields value.
# - `json` : parse a log line using a json format to capture and map fields value.
parsers: #(required)
    # Json parser example
    -   name: "example_json" # ID of the parser, used for alerting and storage (required, must be unique)
        mode: "json" # enable json mode (required for json parser)
        # You must specify in "json_fields" the fields you want to capture (required for json parser)
        json_fields: # mapping between localField to (required for json parser)
            myLocalField: "my_field_in_json"
            date: "datetime"
            level: "level_name"
            message: "message"
            user: "extra.user" # separate fields by "." to capture inner fields

        # File list to include (required)
        # can contains "*" to match pattern or "**" to match all files.  
        files_included:
            - "/var/log/symfony/*.log"
            - "/var/log/**/*.log"
        # File list to exclude (optional)
        # can contains "*" to match pattern or "**" to match all files.
        files_excluded:
            - "/var/log/symfony/prod.deprecations.log"

    # Regex parser example
    -   name: "example_regex" # ID of the parser, used for alerting and storage (required, must be unique)
        mode: "regex"  # enable regex mode (required for regex parser)
        regex_pattern: "\\[(?P<date>.+)\\] [a-zA-Z0-9_\\-]+.(?P<level>[a-zA-Z0-9]+): (?P<message>.*)"  # regex pattern with group name which identify fields to capture (required for regex parser)
        # File list to include (required)
        # can contains "*" to match pattern or "**" to match all files.  
        files_included:
            - "/var/log/symfony/*.log"
            - "/var/log/**/*.log"
        # File list to exclude (optional)
        # can contains "*" to match pattern or "**" to match all files.
        files_excluded:
            - "/var/log/symfony/prod.deprecations.log"

#
# Alerts are used to send notifications to users.
alerts: # (optional)
    frequency: 5 # frequency in seconds to check alerts (optional)
    # Triggers define conditions that must be met to send an alert.
    triggers:
        -   name: "critical detected" # ID of the trigger, displayed in notification (required)
            # List of conditions that must be VALID to trigger (required)
            # "field" must contain the name of a field captured by the parser or a special field from the following list :
            # - "_parser" : name of used parser
            # - "_filename" : filename where current log is found
            # "operator" must contain one of the following operators :
            # - "is" : if field is equal to value (no case sensitive)
            # - "is_not" : if field is not equal to value (no case sensitive)
            # - "contains" : if field contains value (no case sensitive)
            # - "not_contains" : if field not contains value (no case sensitive)
            # - "start_with" : if field start with value (no case sensitive)
            # - "not_start_with" : if field not start with value (no case sensitive)
            # - "match_regex" :  if field match to pattern
            values:
                - { field: "_parser", operator: "is", value: "example_json" }
                - { field: "_filename", operator: "is_not", value: "/var/log/symfony/dev.log" }
                - { field: "level", operator: "is", value: "CRITICAL" }
                - { field: "level", operator: "contains", value: "TICAL" }
                - { field: "level", operator: "not_contains", value: "INFO" }
                - { field: "level", operator: "start_with", value: "CRIT" }
                - { field: "level", operator: "not_start_with", value: "WARN" }
                - { field: "message", operator: "match_regex", value: ".*Error.*" }
    # List of recipients to send notifications to
    # "kind" must contain one of the following types :
    # - "email" : send notification using smtp to an email address
    # - "slack_webhooks" : send slack message using a webhook
    recipients: # (required)
#        -   kind: "email" # (required)
#            recipient: "user@localhost.local"
#        -   kind: "slack_webhook" # (required)
#            recipient: "https://hooks.slack.com/services/XXXX/XXXXX/XXXXX"

# SMTP configuration (optional but required for email notification)
smtp:
    host: "localhost"
    port: 25
    username: ""
    password: ""
    ssl_enabled: false # Is TLS enable (optional)
    from_email: "spooter@localhost" # Email of the email sender (required)
    from_name: "Spooter" # Name of the email sender (required)
```

Next, run `spooter-agent` using:

```bash
################
#   Makefile   #
################
make start-agent config=config_agent.yaml

################
# Command line #
################
# Build
make build-agent
# Run
./bin/spooter-agent -config=config_agent.yaml
```

## Server mode

First, you must create a configuration file `config_server.yml`.

```yaml
#
# Server configuration
#
debug: false

# SMTP configuration (optional but required for email notification)
smtp:
    host: "localhost"
    port: 25
    username: ""
    password: ""
    ssl_enabled: false # Is TLS enable (optional)
    from_email: "noreply@localhost.local" # Email of the email sender (required)
    from_name: "Spooter Bot" # Name of the email sender (required)
```

Next, run `spooter-server` using:

```bash
################
#   Makefile   #
################
make start-aservergent config=config_server.yaml

################
# Command line #
################
# Build
make build-agent
# Run
./bin/spooter-server -config=config_server.yaml
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
