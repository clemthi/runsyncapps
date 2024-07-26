# RunSyncApps

This small application allows to start multiple programs at once and if one of them is closed, the others are killed. Tested on Windows 11 and MacOS 14.5

- [1. Building](#1-building)
- [2. Usage](#2-usage)
- [3. JSON configuration](#3-json-configuration)
- [4. SystemTray icon](#4-systemtray-icon)

## 1. Building

```shell
go build -o="runsyncapps.exe" ./cmd/
```

On Windows, You can add the build flag `-ldflags="-H windowsgui"` to avoid the console to open when starting the app.

## 2. Usage

```shell
runsyncapps.exe --config=myconfig.json --log
```

- `config` : path of the config file (`config.json` by default)
- `log` : log events in a `trace_<timestamp>.log` file (disabled by default)

## 3. JSON configuration

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

## 4. SystemTray icon

A system tray icon allows the using to exit the application without killing the child processes.

This works by default on Windows.

On MacOS you will need to bundle the application (see [getlantern/systray lib info](https://github.com/getlantern/systray/blob/master/README.md) and [official Apple documentation](https://developer.apple.com/library/archive/documentation/CoreFoundation/Conceptual/CFBundles/BundleTypes/BundleTypes.html#//apple_ref/doc/uid/10000123i-CH101-SW1)).
