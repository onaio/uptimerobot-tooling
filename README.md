# UPTIME ROBOT TOOLING

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)  [![Go](https://github.com/bennsimon/uptime-robot-tooling/actions/workflows/go.yaml/badge.svg?branch=main)](https://github.com/bennsimon/uptime-robot-tooling/actions/workflows/go.yaml)

This a tool that helps automate the CRUD tasks for [uptime robot](https://uptimerobot.com/).

## Build from source

* Clone the repository

* From the root of the repo build

  ```shell
  make build
  ```

* `uptimerobot-tooling` executable file should be created on the root.

## Usage

> Rationale for the tool is that `friendly_name` should be unique however this can be disabled by
> setting `MONITOR_RESOLVE_BY_FRIENDLY_NAME` to false(default is true) and
> maintaining `MONITOR_RESOLVE_ALERT_CONTACTS_BY_FRIENDLY_NAME` to false as is default.

For the fields below use the supported values as they will be mapped to their `id`.

| Resource  | Field               | Supported Values                                   |
|-----------|---------------------|----------------------------------------------------|
| `monitor` | `type`              | `HTTP`,`HTTPS`,`Heartbeat`,`Ping`,`Port`,`Keyword` |
| `monitor` | `sub_type`          | `HTTP`,`HTTPS`,`FTP`,`SMTP`,`POP3`,`IMAP`          |
| `monitor` | `keyword_type`      | `exists`, `not exists`                             |
| `monitor` | `keyword_case_type` | `case sensitive`, `case insensitive`               |
| `monitor` | `keyword_case_type` | `Basic`, `Digest`,`HTTP Basic Auth`                |

Arguments Supported:

| Arg | Description                                                                                                                                                                                                                                                                                                                                                                                       | Default Value | Supported Values                           |
|-----|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|--------------------------------------------|
| d   | Data/input                                                                                                                                                                                                                                                                                                                                                                                        | `""`          | text or json (single object or array) file |
| r   | Resource to be acted upon.                                                                                                                                                                                                                                                                                                                                                                        | `monitor`     | `monitor`                                  |
| a   | action to be performed on the resource. Create will create the monitor. <br/>Update will either create or update only if either `id` or `friendly_name` are provided on the payload. Delete removes the monitor using `id` or `friendly_name` to identify a monitor.<br/>In update and delete `id` is given priority over `friendly_name` if both are specified.i.e (update by id, delete by id). | `create`      | `create`, `update`, `delete`               |

Environment Variables Supported:

| Variable                                          | Resource  | Description                                                                                                                                                                                  | Default                           |
|---------------------------------------------------|-----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------|
| `UPTIME_ROBOT_API_KEY`                            | `all`     | Your Uptime robot API key.                                                                                                                                                                   |                                   |
| `UPTIME_ROBOT_API_URL`                            | `all`     | Unless specified otherwise it defaults to https://api.uptimerobot.com/v2/.                                                                                                                   | `https://api.uptimerobot.com/v2/` |
| `MONITOR_RESOLVE_BY_FRIENDLY_NAME`                | `monitor` | If `false` it will not resolve monitor by `friendly_name` i.e updates/deletes will need `id`.                                                                                                | `true`                            |
| `MONITOR_RESOLVE_ALERT_CONTACTS_BY_FRIENDLY_NAME` | `monitor` | if `true` alert contacts can be resolved by their `friendly_name` in addition to `id` i.e instead of supplying its `id` in the alert\_contacts field one can simply use its `friendly_name`. | `false`                           |
| `MONITOR_ALERT_CONTACTS_DELIMITER`                | `monitor` | Delimiter used to separate alert contacts when creating/updating a monitor. Default as specified [here](https://uptimerobot.com/api/).                                                       | `-`                               |
| `MONITOR_ALERT_CONTACTS_ATTRIB_DELIMITER`         | `monitor` | Delimiter used to separate alert contacts attributes when creating/updating a monitor. Default as specified [here](https://uptimerobot.com/api/).                                            | `_`                               |

### Examples

(Running from the root of the repository)

### Help

```shell
./uptimerobot-tooling -h
```

### Using json file

```json
{
  "friendly_name": "example-com",
  "url": "https://example.com",
  "type": "HTTP"
}
```

```shell
./uptimerobot-tooling -r=monitor -a=update -d=test.json
```

### Using string

```shell
./uptimerobot-tooling -r=monitor -a=delete -d='{"friendly_name":"example-com","url":"https://example.com","type":"HTTP"}'
```

or

```shell
./uptimerobot-tooling -r=monitor -a=delete -d='{"id":"34","url":"https://example.com","type":"HTTP"}'
```

### Specifying alert contacts when working on monitors

* If `MONITOR_RESOLVE_ALERT_CONTACTS_BY_FRIENDLY_NAME` is `false` on your setup ignore this section.
* From the api [documentation](https://uptimerobot.com/api/) one can specify alert contact with or without appending
  recurrence and threshold attributes like so `1`, `1_0_6`. With this tooling one can use either `friendly_name`
  or `id` of alert contact like
  so `alertContactA`,`1`, `alertContactA_0_6`, `alertContactA_0_6-alertContactA_0-5`, `1_0_6`.

      {
        ...
        "alert_contacts": "alertContactA_0_6"
      }
      ...
      [{
        ...
        "alert_contacts": "1_0_6"
      }]
* If alert contact `friendly_name` has `-` or `_` one can override the default delimiters for alert contacts and alert
  contacts attributes by specifying `MONITOR_ALERT_CONTACTS_DELIMITER` and/or `MONITOR_ALERT_CONTACTS_ATTRIB_DELIMITER`
  respectively. For example instead of:
    * `alert-contact-a_0_5-|alert-contact-b` (friendly\_name has hyphen that will affect alert contact splitting)
      set `MONITOR_ALERT_CONTACTS_DELIMITER` to `|` then
      supply `alert-contact-a_0_5|alert-contact-b`.
    * `alert_contact_a|0|5`(friendly\_name has underscore that will affect alert contact attribute splitting)
      set `MONITOR_ALERT_CONTACTS_ATTRIB_DELIMITER`
      to `|` and `MONITOR_ALERT_CONTACTS_DELIMITER` to `-`
      then supply `alert_contact_a|0|5-alert_contact_b|0|5`.
