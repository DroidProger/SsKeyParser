# SsKeyParser
Parser for shadowsocks keys on web pages

This app is extract shadowsocks keys from web pages and telegramm channels
To run it type in terminal "path_to_app path_to_config_file"
Example for linux
```
/bin/sskeyparser-nix64 /etc/sskeyparser/config.json
```
### Parameters explanation
```
"ssconfigfile":"/etc/shadowsocks/config.json",
```
Path to shadowsocks config file
```
"sspath":"/bin/systemctl",
"ssrestartcommand":[
  "restart",
  "sslocal.service"
],
```
Parameter block for restarting shadowsocks. If shadowsocks is not running as a service, this section may look like this.
```
"sspath":"/bin/shadowsocks/sslocal",
    "ssrestartcommand":[
        "-c",
        "/etc/shadowsocks/config.json"
    ],
```

```
"ssconfigsectionpath":[
        "servers"
    ],
```
Section name in shadowsocks config file, where servers will be added. Your shadowsocks config file must contain this section even it has no servers configurations
Exampl of shadowsocks config.json
```
{
  "local_adress"
}
```
