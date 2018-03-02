# mydns-notifier

## install(systemd)
```sh
sudo bash -c "$(curl -fsS https://raw.githubusercontent.com/hrntknr/mydns-notifier/master/install_systemd.sh)"
```

## usage
```
Usage of mydnsnotifier:
  -config string
        configPath (default "./config.toml")
  -verbosity int
        verbosity, 1,2,3,4,5 (default 3)
```

## default config
```toml
[notice]
id = ""
password = ""
ipv4 = true
ipv6 = true
cron = ""

[log.slack]
enable = false
hookURL = ""
```
