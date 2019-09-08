[![Build Status](https://travis-ci.org/oleg-balunenko/instadiff-cli.svg?branch=master)](https://travis-ci.org/oleg-balunenko/instadiff-cli)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=instadiff-cli&metric=alert_status)](https://sonarcloud.io/dashboard?id=instadiff-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/oleg-balunenko/instadiff-cli)](https://goreportcard.com/report/github.com/oleg-balunenko/instadiff-cli)
[![Coverage Status](https://coveralls.io/repos/github/oleg-balunenko/instadiff-cli/badge.svg?branch=dev)](https://coveralls.io/github/oleg-balunenko/instadiff-cli?branch=dev)
[![Go Report Card](https://goreportcard.com/badge/github.com/oleg-balunenko/instadiff-cli)](https://goreportcard.com/report/github.com/oleg-balunenko/instadiff-cli)
[![Latest release artifacts](https://img.shields.io/badge/artifacts-download-blue.svg)](https://github.com/oleg-balunenko/instadiff-cli/releases/latest)
# instadiff-cli

instadiff-cli - a command line tool for managing instagram account followers and followings

## Usage:

```shell script
instadiff-cli help
```

```text

NAME:
   instadiff-cli - a command line tool for managing instagram account followers and followings

USAGE:
   instadiff-cli [global options] command [command options] [arguments...]


COMMANDS:
   list-followers, followers        list your followers
   list-followings, followings      list your followings
   clean-followers, clean           Un follow not mutual followings, except of whitelisted
   not-mutual-followings, unmutual  List all not mutual followings
   help, h                          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config_path value  Path to the config file (default: ".config.json")
   --help, -h           show help
   --version, -v        print the version
```

To get help for any supported command:
``` shell script
instadiff-cli help [command]
```

Example of config file:

```json
{
  "user":{
    "instagram":{
      "username":"user",
      "password":"pass"
    }
  },
  "whitelist":[
    "user1",
    "user2",
    "user3"
  ],
  "limits":{
    "unfollow":100,
    "follow": 50
  },
  "debug": "false"
}
```

Create a json file with configuration and pass the path to it via flag `--config_path`
```shell script
instadiff-cli --config_path ".config.json" [command]
```
