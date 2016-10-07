![karmabot logo](/logo.png)

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
    1. `go get`
    1. `go build`

### Download a Pre-built Release

1. head to [the repo's releases page](https://github.com/kamaln7/karmabot/releases) and download the appropriate latest release's binary for your system

## Usage

1. add a **Slack Bot** integration: `https://team.slack.com/apps/A0F7YS25R-bots` 
2. invite `karmabot` to any existing channels and all future channels (this is a limitation of Slack's bot API, unfortunately)
3. run `karmabot`. the following options are supported:


| option                  | required? | description                              | default        |
| ----------------------- | --------- | ---------------------------------------- | -------------- |
| `-token string`         | **yes**   | slack RTM token                          |                |
| `-debug bool`        | no        | set debug mode | `false`            |
| `-db string`            | no        | path to sqlite database                  | `./db.sqlite3` |
| `-leaderboardlimit int` | no        | the default amount of users to list in the leaderboard | `10`           |
| `-maxpoints int`        | no        | the maximum amount of points that users can give/take at once | `6`            |

**example:** `./karmabot -token xoxb-abcdefg`

## License

see [./LICENSE](/LICENSE)
