![karmabot logo](/logo.png)  

[![Build Status](https://semaphoreci.com/api/v1/kamaln7/karmabot/branches/master/badge.svg)](https://semaphoreci.com/kamaln7/karmabot)

karmabot is a Slack bot that listens for and performs karma operations (aka upvotes/downvotes).

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
- [motivate.im](http://motivate.im/) support:
  - `?m <user>`
  - `!m <user>`
- leaderboard:
  - `<karma|karmabot> <leaderboard|top|highscores>`
  - to list more than `leaderboardlimit` (see the **Usage** section below), you may append the number of users to list to the command above. e.g. `karmabot top 20`


**note:** `<user>` does not have to be a Slack username. However, karmabot supports Slack autocompletion and so the following messages are parsed correctly:

- `@username: ++`
- `@username++`
- `username: ++`
- `!m @username: `
- etc.

## Installation

### Build from Source

1. clone the repo:
    1. `git clone -b v0.1.0 https://github.com/kamaln7/karmabot.git`
2. run `go get` and then `go build` inside the repo's root
    1. `cd karmabot`
    2. `go get`
    3. `go build`

### Download a Pre-built Release

1. head to [the repo's releases page](https://github.com/kamaln7/karmabot/releases) and download the appropriate latest release's binary for your system

## Usage

1. add a **Slack Bot** integration: `https://team.slack.com/apps/A0F7YS25R-bots` 
2. invite `karmabot` to any existing channels and all future channels (this is a limitation of Slack's bot API, unfortunately)
3. run `karmabot`. the following options are supported:


| option                  | required? | description                              | default        |
| ----------------------- | --------- | ---------------------------------------- | -------------- |
| `-token string`         | **yes**   | slack RTM token                          |                |
| `-debug bool`           | no        | set debug mode                           | `false`        |
| `-db string`            | no        | path to sqlite database                  | `./db.sqlite3` |
| `-leaderboardlimit int` | no        | the default amount of users to list in the leaderboard | `10`           |
| `-maxpoints int`        | no        | the maximum amount of points that users can give/take at once | `6`            |
| `-motivate bool`        | no        | toggle [motivate.im](http://motivate.im/) support | `true`         |

In addition, see the table below for the options related to the web UI.

**example:** `./karmabot -token xoxb-abcdefg`

## Web UI

karmabot includes an optional web UI. The web UI uses TOTP tokens for authentication. While the token itself would only be valid for 30 seconds, once you have authenticated, you will stay so for 48 hours, after which your session will expire. This is not meant to be a fully-featured advanced authentication system, but rather a simple way to keep off people who do not belong to your Slack team.

### How to use the Web UI

#### Requisites

1. download the `www` directory from the repo's root and place it in a directory that is accessible to karmabot.
2. run `./karmabot -token YOUR_SLACK_TOKEN -webuipath /path/to/www -listenaddr 127.0.0.1:9000`. The initial values do not matter, as they will not be used at all. karmabot will generate a random TOTP key for you to use, print it, and exit. Copy that token.

#### Start karmabot

Once you have performed the steps detailed above, pass the necessary options to the `karmabot` binary:

| option               | required? | description                              | default                           |
| -------------------- | --------- | ---------------------------------------- | --------------------------------- |
| `-listenaddr string` | **yes**   | the address (`host:port`) on which to serve the web UI |                                   |
| `-totp string`       | **yes**   | the TOTP key (see above)                 |                                   |
| `-webuipath string`  | **yes**   | path to the `www` directory (see above)  |                                   |
| `-webuiurl string`   | no        | the URL which karmabot should use to generate links to the web UI | defaults to `-listenaddr`'s value |


If done correctly, the web UI should be accessible on the `listenaddr` that you have configured.

#### Usage

The web UI is authenticated, so you will have to generate authentication tokens through karmabot. You can access the web UI by typing `karmabot web` in the chat. karmabot will generate a TOTP token, append it to the `webuiurl` and send back the link. Click on the link and you should be authenticated for 48 hours.

Additionally, you may use also use the link provided in the Slack leaderboard (`karmabot leaderboard`) in order to log in and access the leaderboard.

## License

see [./LICENSE](/LICENSE)
