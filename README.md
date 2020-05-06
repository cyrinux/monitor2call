# Monitor2Call

A future API notifications proxy

/!\ WIP (but working)

## Feature

Should be able to notify user based on subscribed tags

### Supported notifications way

- Pushover (with alerts acknowlegement)
- Mail
- Phone call
- SMS

### Others

- Escalade support

## Requirements

- put json google api keys in `./keys/google_api_keys.json`
- export pushover app api key with `export PUSHOVER_AP_API_KEY=yourappapikey`
- you can create a ".env" file in this folder with all VAR needed (see head of)

## Usage

### Build and run in local

make

### Build and run release in docker

make WRITE_PASSWORD=WriteMonitor2Call READ_PASSWORD=ReadMonitor2Call PUSHOVER_AP_API_KEY=mypushoverapikey compose
make up && make logs
