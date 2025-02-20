# badcfg

badcfg is a dead simple tool to list and copy BAD config file values

## Usage:

- `list [key]`: List all keys in the config file matching the given key
- `copy <key>`: Copy a value from the config file matching the given key (exact
  match)

## Examples:

| Command                                       | Description                                                     |
| --------------------------------------------- | --------------------------------------------------------------- |
| `badcfg list`                                 | List all keys in the config file                                |
| `badcfg list jetprofile.datasource.dev1`      | List all keys matching jetprofile.datasource.dev1.*             |
| `badcfg copy jetprofile.datasource.dev1.host` | Copy the value of jetprofile.datasource.dev1.host               |
| `badcfg copy jetprofile`                      | This will result in an error, because more than one key matches |
