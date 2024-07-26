# RunSyncApps

This small application allows to start multiple programs at once and if one of them is closed, the others are killed. A system tray icon allows the using to exit the application without killing the child processes.

## Bulding

```shell
go build -o="runsyncapps.exe" ./cmd/
```

You can add the build flag `-ldflags="-H windowsgui"` to avoid the console to open when starting the app.

## Usage

```shell
runsyncapps.exe --config=myconfig.json --log
```

- `config` : path of the config file (`config.json` by default)
- `log` : log events in a `trace_<timestamp>.log` file (disabled by default)

## JSON configuration

The configuration file looks like this:

```json
{
    "waitCheck": 10,
    "waitExit": 10,
    "applications": [
        {
            "path": "C:\\Windows\\System32\\dxdiag.exe",
            "useExistingInstance": false,
            "killOnExit": true
        },
        {
            "path": "C:\\Windows\\System32\\charmap.exe",
            "useExistingInstance": false,
            "killOnExit": true
        },
        {
            "path": "C:\\Windows\\System32\\msinfo32.exe",
            "useExistingInstance": false,
            "killOnExit": false
        }
    ]
}
```

The parameters are the following:

- `waitCheck`: timer (in seconds) after the processes are being monitored
- `waitExit`: timer (in seconds) before the processes are being killed
- `applications`: array of application to start:
  - `path`: full path of the application
  - `useExistingInstance`: don't start a new instance if there is one already running
  - `killOnExit`: kill the application if it's running after another app has been killed
