- Author: xiongjiwei <xiongjiwei1996@outlook.com>
- Start Date: 2021-08-29
- RFC PR: N/A
- Tracking Issue: [beyondstorage/beyond-ftp#12](https://github.com/beyondstorage/beyond-ftp/issues/12)

# RFC-1: Command Line Design

## Background

There are many different configs can be set in Beyond-FTP, such as storager type, host, and port. It is useful to design a command line interface to make some of them easy to configurable.

## Proposal

So I propose the following commands:

| name    | shorthand | default      | usage                      |
|---------|:---------:|--------------|----------------------------|
| version | v         |              | show beyond-ftp version    |
| help    | h         |              | show usage of beyond-ftp   |
| config  | c         |              | config file path           |
| host    |           | 127.0.0.1    | server listen host         |
| port    | p         | 21           | server listen port         |
| debug   | d         | false        | start with debug mode      |


## Rationale

`config` provide a config file to start FTP server, if no file is specified, the configures will use its default value.

`host` and `port` provide the host and port that the server should listend. The default value is `127.0.0.1:21`.

`debug` indicate the server start with debug mode. Debug mode will print log with debug level and provide profiles, and use memory as under storage.

The other configure can be found in `config/config.example.toml`.

Server will first use the command line specifie value, if no specified, use the value in config file, then, the default value.


## Compatibility

N/A

## Implementation

N/A