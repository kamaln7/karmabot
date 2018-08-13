![karmabot logo](/logo.png)  

[![Build Status](https://semaphoreci.com/api/v1/kamaln7/karmabot/branches/master/badge.svg)](https://semaphoreci.com/kamaln7/karmabot)

karmabot is a Slack bot that listens for and performs karma operations (aka upvotes/downvotes).

<img src="/screenshot.png" alt="usage screenshot" width="445">

## Syntax

- upvote a user: `<user>++`
- downvote a user: `<user>--`
- add/subtract multiple points at once:
  - `<user>++` - 1 point
  - `<user>+++` - 2 points
  - and so on, limited to `maxpoints` points (see the **Usage** section below)
- add a message/reason for a karma operation:
  - `<user>++ for <message>`; or
  - `<user>++ <message>`
- query a user's current points: `<user>==`
- upvote/downvote a user by adding reactjis to their message
- [motivate.im](http://motivate.im/) support:
  - `?m <user>`
  - `!m <user>`
- leaderboard:
  - `<karma|karmabot> <leaderboard|top|highscores>`
  - to list more than `leaderboardlimit` (see the **Usage** section below), you may append the number of users to list to the command above. e.g. `karmabot top 20`
- user aliases:
  - it is possible to alias different usernames to one main username by passing the aliases as a cli option to the karmabot binary. syntax: `-alias main++alias1++alias2++...++aliasN`
  - repeat the option for every alias that you want to configure
- karma throwback:
  - `<karma|karmabot> throwback [user]`
  - returns a random karma operation that happened to a specific user.

**note:** `<user>` does not have to be a Slack username. However, karmabot supports Slack autocompletion and so the following messages are parsed correctly:

- `@username: ++`
- `@username++`
- `username ++`
- `!m @username: `
- etc.

## Installation

### Build from Source

1. clone the repo:
    1. `git clone -b v1.5.0 https://github.com/kamaln7/karmabot.git`
2. run `go get` and then `go build` in `/cmd/karmabot` and `/cmd/karmabotctl`
    1. `cd karmabot`
    2. `go get`
    3. `cd cmd/karmabot`
    4. `go build`
    5. `cd ../karmabotctl`
    6. `go build`

### Download a Pre-built Release

1. head to [the repo's releases page](https://github.com/kamaln7/karmabot/releases) and download the appropriate latest release's binary for your system

## Usage

1. add a **Slack Bot** integration: `https://team.slack.com/apps/A0F7YS25R-bots`. an avatar is available [here](/avatar.png).
2. invite `karmabot` to any existing channels and all future channels (this is a limitation of Slack's bot API, unfortunately)
3. run `karmabot`. the following options are supported. you can use environment variables as well, but any CLI options you pass will take precedence.


| option                      | required? | description                                                  | default                          | env var                |
| --------------------------- | --------- | ------------------------------------------------------------ | -------------------------------- | ---------------------- |
| `-token string`             | **yes**   | slack RTM token                                              |                                  | `KB_TOKEN`             |
| `-debug=bool`               | no        | set debug mode                                               | `false`                          | `KB_DEBUG`             |
| `-db string`                | no        | path to sqlite database                                      | `./db.sqlite3`                   | `KB_DB`                |
| `-leaderboardlimit int`     | no        | the default amount of users to list in the leaderboard       | `10`                             | `KB_LEADERBOARDLIMIT`  |
| `-maxpoints int`            | no        | the maximum amount of points that users can give/take at once | `6`                              | `KB_MAXPOINTS`         |
| `-motivate=bool`            | no        | toggle [motivate.im](http://motivate.im/) support            | `true`                           | `KB_MOTIVATE`          |
| `-blacklist string`         | no        | **may be passed multiple times** blacklist `string`  i.e. ignore karma commands for `string` | `[]`                             | `KB_BLACKLIST`         |
| `-reactji bool`             | no        | use reactji (👍 and 👎) as reaction events                     | `true`                           | `KB_REACTJI`           |
| `-reactjis.upvote string`   | no        | **may be passed multiple times** a list of reactjis to use for upvotes. for emojis with aliases, use the first name that is shown in the emoji popup | `+1`, `thumbsup`, `thumbsup_all` | `KB_REACTJIS_UPVOTE`   |
| `-reactjis.downvote string` | no        | **may be passed multiple times** a list of reactjis to use for downvotes. for emojis with aliases, use the first name that is shown in the emoji popup | `-1`, `thumbsdown`               | `KB_REACTJIS_DOWNVOTE` |
| `-alias string`             | no        | **may be passed multiple times** alias different users to one user. syntax: `-alias main++alias1++alias2++...++aliasN` |                                  | `KB_ALIAS`             |
| `-selfkarma bool`           | yes       | allow users to add/remove karma to themselves                | `true`                           | `KB_SELFKARMA`         |

In addition, see the table below for the options related to the web UI.

**example:** `./karmabot -token xoxb-abcdefg`

It is recommended to pass karmabot's logs through [humanlog](https://github.com/aybabtme/humanlog). humanlog will format and color the JSON output as nice easy-to-read text.

## Web UI

karmabot includes an optional web UI. The web UI uses TOTP tokens for authentication. While the token itself would only be valid for 30 seconds, once you have authenticated, you will stay so for 48 hours, after which your session will expire. This is not meant to be a fully-featured advanced authentication system, but rather a simple way to keep off people who do not belong to your Slack team.

### How to use the Web UI

#### Requisites

1. download the `www` directory from the repo's root and place it in a directory that is accessible to karmabot.
2. run `./karmabot -token x -webui.listenaddr x -webui.path x`. You may keep all the options set to `x`, as they will not be used at all. karmabot will generate a random TOTP key for you to use, print it, and exit. Copy that token.

#### Start karmabot

Once you have performed the steps detailed above, pass the necessary options to the `karmabot` binary. You can use environment variables as well, but any CLI options you pass will take precedence.

| option                     | required? | description                                                  | default                               | env var               |
| -------------------------- | --------- | ------------------------------------------------------------ | ------------------------------------- | --------------------- |
| `-webui.listenaddr string` | **yes**   | the address (`host:port`) on which to serve the web UI       |                                       | `KB_WEBUI_LISTENADDR` |
| `-webui.totp string`       | **yes**   | the TOTP key (see above)                                     |                                       | `KB_WEBUI_TOTP`       |
| `-webui.path string`       | **yes**   | path to the `www` directory (see above)                      |                                       | `KB_WEBUI_PATH`       |
| `-webui.url string`        | no        | the URL which karmabot should use to generate links to the web UI (_without_ a trailing slash!) | defaults to `http://webui.listenaddr` | `KB_WEBUI_URL`        |


If done correctly, the web UI should be accessible on the `webui.listenaddr` that you have configured. The web UI will not be started if either of `webui.listenaddr` or `webui.path` are missing.

#### Usage

The web UI is authenticated, so you will have to generate authentication tokens through karmabot. You can access the web UI by typing `karmabot web` in the chat. karmabot will generate a TOTP token, append it to the `webuiurl` and send back the link. Click on the link and you should be authenticated for 48 hours.

Additionally, you may use also use the link provided in the Slack leaderboard (`karmabot leaderboard`) in order to log in and access the leaderboard.

## karmabotctl

karmabot comes with a maintenance tool called `karmabotctl`. It can be used to perform certain tasks without having to run `karmabot` itself.

### Commands

A list of all arguments for each command can be printed by running `karmabotctl karma migrate --help`. In addition to the arguments listed in the tables below, some commands may also require a `<db>` argument containing the path to the database file.

#### karma

| command   | arguments                       | description                             |
| --------- | ------------------------------- | --------------------------------------- |
| add       | `<from> <to> <reason> <points>` | add karma to a user                     |
| migrate   | `<from> <to>`                   | move a user's karma to another user     |
| reset     | `<user>`                        | reset a user's karma                    |
| set       | `<user> <points>`               | set a user's karma to a specific number |
| throwback | `<user>`                        | get a karma throwback for a user        |

#### webui

| command | arguments                                | description                              |
| ------- | ---------------------------------------- | ---------------------------------------- |
| serve   | `<debug> <leaderboardlimit> <totp> <path> <listenaddr> <url>` | start a webserver                        |
| totp    | `<totp>`                                 | generate a TOTP token based on the passed secret |

## License

see [./LICENSE](/LICENSE)
