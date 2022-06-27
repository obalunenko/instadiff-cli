![coverbadger-tag-do-not-edit](https://img.shields.io/badge/coverage-12.81%25-brightgreen?longCache=true&style=flat)

[![GO](https://img.shields.io/github/go-mod/go-version/obalunenko/instadiff-cli)](https://golang.org/doc/devel/release.html)
[![Build Status](https://travis-ci.com/obalunenko/instadiff-cli.svg?branch=master)](https://travis-ci.com/obalunenko/instadiff-cli)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=obalunenko_instadiff-cli&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=obalunenko_instadiff-cli)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/e1b08a94c9cb45f4ac86391ef936166e)](https://www.codacy.com/manual/oleg.balunenko/instadiff-cli?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=obalunenko/instadiff-cli&amp;utm_campaign=Badge_Grade)
[![Go Report Card](https://goreportcard.com/badge/github.com/obalunenko/instadiff-cli)](https://goreportcard.com/report/github.com/obalunenko/instadiff-cli)
[![Latest release artifacts](https://img.shields.io/github/v/release/obalunenko/instadiff-cli)](https://github.com/obalunenko/instadiff-cli/releases/latest)
[![License](https://img.shields.io/github/license/obalunenko/instadiff-cli)](/LICENSE)

# instadiff-cli

<p align="center">
  <img src="https://github.com/obalunenko/instadiff-cli/blob/master/assets/gopher.png" alt="" width="300">
  <br>
</p>

instadiff-cli - a command line tool for managing instagram account followers and followings

## Usage

```shell script
instadiff-cli help
```

```text


██╗███╗   ██╗███████╗████████╗ █████╗ ██████╗ ██╗███████╗███████╗     ██████╗██╗     ██╗
██║████╗  ██║██╔════╝╚══██╔══╝██╔══██╗██╔══██╗██║██╔════╝██╔════╝    ██╔════╝██║     ██║
██║██╔██╗ ██║███████╗   ██║   ███████║██║  ██║██║█████╗  █████╗█████╗██║     ██║     ██║
██║██║╚██╗██║╚════██║   ██║   ██╔══██║██║  ██║██║██╔══╝  ██╔══╝╚════╝██║     ██║     ██║
██║██║ ╚████║███████║   ██║   ██║  ██║██████╔╝██║██║     ██║         ╚██████╗███████╗██║
╚═╝╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═════╝ ╚═╝╚═╝     ╚═╝          ╚═════╝╚══════╝╚═╝


NAME:
   instadiff-cli - a command line tool for managing instagram account followers and followings

USAGE:
   instadiff-cli [global options] command [command options] [arguments...]

VERSION:

| app_name:     instadiff-cli                            |
| version:      v1.6.0                                   |
| go_version:   go1.18.3                                 |
| commit:       e03384f8050cd48b3c824a562c248ed952333f4d |
| short_commit: e03384f8                                 |
| build_date:   2022-06-27T15:59:12Z                     |
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||


AUTHOR:
   Oleg Balunenko <oleg.balunenko@gmail.com>

COMMANDS:
   list-followers, followers                                                   List your followers
   list-followings, followings                                                 List your followings
   clean-followings, clean, unfollow-untmutual, remove-untmutual, rm-unmutual  Un follow not mutual followings, except of whitelisted
   remove-followers, rm, remove                                                Remove a list of followers, by username.
   unfollow-users, unfollow, remove-followings                                 Unfollow a list of followings, by username.
   follow-users, follow, add-followings                                        Follow a list of followings, by username.
   list-unmutual, unmutual                                                     List all not mutual followings
   list-useless, useless, bots                                                 List all statistic-useless accounts (bots, business accounts or mass-followers) (alpha)
   list-diff, diff                                                             List diff for account (lost and new followers and followings)
   diff-history, history                                                       List diff account history (lost and new followers and followings)
   help, h                                                                     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config_path value  Path to the config file (default: ".config.json")
   --help, -h           show help (default: false)
   --incognito          Incognito removes session on application exit. (default: false)
   --log_level value    Level of output logs (default: "INFO")
   --version, -v        print the version (default: false)
   
```

To get help for any supported command:

``` shell script
instadiff-cli help [command]
```

Example of config file:

```json
{
  "instagram":{
    "whitelist":[
      "user1",
      "user2",
      "user3"
    ],
    "limits":{
      "unfollow":100
    },
    "sleep": 1
  },
  "storage": {
    "local": true,
    "mongo": {
      "url": "mongoURL:test",
      "db": "testing"
    }
  }
}
```

* instagram: it is a config for instagram
    * whitelist: list of followings that will be not unfollowed even if they are not mutual (usernames and ID's supported both).
    * limits: limits per one run.
        * unfollow: number of users that could be unfollowed in one run (be careful with big number - account could be banned)
    * sleep: sleep interval in seconds between each unfollow request to prevent account ban for ddos reason.
* storage: it's a config for database storage. 
	* local: if true, memory cache will be used and connection to mongo will be not set.
	* mongo: is a config for mongo database
	  - url: url of mongo DB to connect
	  - db: name of Database

Create a json file with configuration and pass the path to it via flag `--config_path`

```shell script
instadiff-cli --config_path ".config.json" [command]
```

## Develop

To start developing - create the fork of repository, make changes and open PR to the origin.

### Build

Run `make build` command in the root of repository to compile the binary and test locally changes.

### Testing

Run `make test` command in the root of repository to execute unit tests.
