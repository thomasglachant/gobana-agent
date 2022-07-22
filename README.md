# spooter

## Configuration

First, you must create a configuration file `config.yml`.

```yaml
#
# Agent configuration
#
agent:
    #
    # Metadata about the agent 
    metadata: # (required)
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
            regex_pattern: "\\[(.+)\\] [a-zA-Z0-9_\\-]+.([a-zA-Z0-9]+): (.*)"  # regex pattern (required for regex parser)
            # List of fields to capture in pattern (required for regex parser)
            regex_fields:
                - "date"
                - "level"
                - "message"
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
        # "type" must contain one of the following types :
        # - "email" : send notification using smtp to an email address
        # - "slack_webhooks" : send slack message using a webhook
        recipients: # (required)
            -   type: "email" # (required)
                recipient: "user@localhost.local"
            -   type: "slack_webhook" # (required)
                recipient: "https://hooks.slack.com/services/XXXX/XXXXX/XXXXX"

#
# Global configuration
#
common:
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

## Build 

```bash 
make build 
```

## Run

```bash
# Using makefile
make start-agent config=config.yaml

# Using command line
./spooter -agent -config=config.yaml
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

## Regex patterns

* Symfony logs regex : `\\[(.+)\\] [a-zA-Z0-9_\\-]+.([a-zA-Z0-9]+): (.*)`
* Nginx access
  regex : `%{IPORHOST:visitor_ip} (?:-|(%{WORD}.%{WORD})) %{USER:ident} \[%{HTTPDATE:time_local}\] "%{METHOD:method} %{URIPATHPARAM:http_path} HTTP/%{NUMBER:http_version}" %{INT:status} %{INT:body_bytes_sent} "(?:-|(%{URI:referer}))" %{QS:user_agent}`
