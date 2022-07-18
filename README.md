# Spotter

## Configuration

First, you must create a configuration file.

```yaml
########################
# Client configuration #
########################
client:
    metadata:
        application: "MyAwesomeApp"
        server: "localhost"

    mode: "standalone"

    lookups:
        -   name: "php-symfony"
            patterns:
                -   name: "criticals"
                    type: "regex"
                    value: "(.*)CRITICAL(.*)"
                -   name: "database"
                    type: "regex"
                    value: "(.*)doctrine(.*)"
            files:
                - "/srv/myapp/var/log/prod.log"
                - "/srv/myapp/var/log/dev.log"
        -   name: "nginx"
            patterns:
                -   name: "error_500"
                    type: "regex"
                    value: "(.*)CRITICAL(.*)"
            files:
                - "/var/log/nginx/*.log"
    
    alerts:
        subscriptions:
            -   type: "email"
                value: "user@localhost.local"
                lookups: [ "symfony" ]
            -   type: "slack"
                value: "https://hooks.slack.com/services/XXXXXXXX/XXXXXXXXX/XXXXXXXX"
                lookups: [ "symfony" ]
                
    smtp:
        host: "localhost"
        port: 25
        username: ""
        password: ""
        ssl_enabled: false
        from_email: "np-reply@spotted.local"
        from_name: "Spotter Bot"
        

```

## Build and run

```bash 
make build 
```

Run using Makefile :
```bash
make start-client config=config.yaml
```

Run using command line :
```bash
./spotter -client -config=config.yaml
```
