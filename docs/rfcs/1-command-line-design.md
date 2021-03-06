- Author: xiongjiwei <xiongjiwei1996@outlook.com>
- Start Date: 2021-08-29
- RFC PR: N/A
- Tracking Issue: [beyondstorage/beyond-ftp#12](https://github.com/beyondstorage/beyond-ftp/issues/12)

# RFC-1: Command Line Design

## Background

Command line is a interface to use the software, user can use it to control the software's behavior.

## Proposal

A command line program include the `commands`, `flags` and `args`. For Beyond-FTP, they are:

- subcommands

| name    | usage                      |
|---------|----------------------------|
| version | show beyond-ftp version    |
| help    | show usage of beyond-ftp   |

These two subcommands does not have any `flags` or `args`.

- flags

| name   | shorthand | default   | type   | require | usage                        |
|--------|-----------|-----------|--------|---------|------------------------------|
| config | c         |           | string | N       | config file path             |
| host   |           | 127.0.0.1 | string | N       | server listen host           |
| port   | p         | 21        | number | N       | server listen port           |
| debug  | d         | false     | bool   | N       | start server with debug mode |

`config` provide a config file to start FTP server, if no file is specified, the configures will use its default value.

`host` and `port` provide the host and port that the server should listend. The default value is `127.0.0.1:21`.

`debug` indicate the server start with debug mode. Debug mode will print log with debug level and provide profiles, and use memory as under storage.

- args

N/A

- config file

use `toml` as the config file format, it is easy for human read. It contains the following fields:
| name        | default        | type   | comment                               |
|-------------|----------------|--------|---------------------------------------|
| storager    | memory://ftp   | string | connection string for Storager init   |
| host        | 127.0.0.1      | string | server listen host                    |
| port        | 21             | number | server listen port                    |
| public-host | 127.0.0.1      | string | server passive connection listen host |
| users       | anonymous = "" |        | server usename and password           |

an example config file will be provided under `config` directory.

- env values
We support the following env values, each of them represent a field in the config file. The value of `BEYOND_FTP_USERS` should be `json` format. 
| env value name      | config      |
|---------------------|-------------|
| BEYOND_FTP_SOTRAGER | storager    |
| BEYOND_FTP_HOST     | host        |
| BEYOND_FTP_PORT     | port        |
| BEYOND_FTP_PUB_HOST | public-host |
| BEYOND_FTP_USERS    | users       |

- apply order

The apply order is `flags`, `config file`, `env values`, `default value`

## Rationale

N/A

## Compatibility

N/A

## Implementation

- use lib https://github.com/urfave/cli to parse the command and flag.
- Beyond-FTP return 0 if normally exit, other for error occur.
