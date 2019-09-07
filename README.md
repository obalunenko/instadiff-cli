# insta-follow-diff

instadiff-cli - a command line tool for managing instagram account followers and followings

## Usage:

`
instadiff-cli help
`


```shell script

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