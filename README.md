# instadiff-cli

[![Logo](.assets/gopher.png)]
[![Build Status](https://travis-ci.org/oleg-balunenko/instadiff-cli.svg?branch=master)](https://travis-ci.org/oleg-balunenko/instadiff-cli)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=instadiff-cli&metric=alert_status)](https://sonarcloud.io/dashboard?id=instadiff-cli)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/e1b08a94c9cb45f4ac86391ef936166e)](https://www.codacy.com/manual/oleg.balunenko/instadiff-cli?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=oleg-balunenko/instadiff-cli&amp;utm_campaign=Badge_Grade)
[![Go Report Card](https://goreportcard.com/badge/github.com/oleg-balunenko/instadiff-cli)](https://goreportcard.com/report/github.com/oleg-balunenko/instadiff-cli)
[![Coverage Status](https://coveralls.io/repos/github/oleg-balunenko/instadiff-cli/badge.svg?branch=master)](https://coveralls.io/github/oleg-balunenko/instadiff-cli?branch=master)
[![Latest release artifacts](https://img.shields.io/badge/artifacts-download-blue.svg)](https://github.com/oleg-balunenko/instadiff-cli/releases/latest)

instadiff-cli - a command line tool for managing instagram account followers and followings

## Usage

```shell script
instadiff-cli help
```

```text

NAME:
   instadiff-cli - a command line tool for managing instagram account followers and followings

USAGE:
   instadiff-cli [global options] command [command options] [arguments...]

COMMANDS:
   followers               List your followers
   followings              List your followings
   clean-followers, clean  Un follow not mutual followings, except of whitelisted
   unmutual, unmutual      List all not mutual followings
   help, h                 Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config_path value  Path to the config file (default: ".config.json")
   --log_level value    Level of output logs (default: "info")
   --debug              Debug mode, where actions has no real effect
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
