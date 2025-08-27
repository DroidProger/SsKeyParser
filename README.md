# SsKeyParser
Parser for shadowsocks keys on web pages

### Always backup your worked shadowsocks config.json 

This app is extract shadowsocks keys from web pages and telegramm channels
To run it type in terminal "path_to_app" "path_to_config_file"
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
Section name in shadowsocks config file, where servers will be added. Your shadowsocks config file must contain this section even it has no servers configurations.

Example of shadowsocks config.json
```
{
  "local_adress":"0.0.0.0"
  ...
  "servers":[
    {
      "server":"permanent_server"
    }
    ...
    {
      "server":"temporary_server"
    }
  ],
  "local_port":1234
  ...
}
```

```
"ssserverseditpos":1,
```
Position from which servers will be edited. In example above first server config will be save, other configs will be rewrite.
```
"sstimeoutdefault":0,
```
Default timeout. Ignored when zero
```
"outputfile":"/etc/sskeyparser/parsingresult.json",
```
Results of parsing is save to this file.
```
"links":[
        {
            "url":"https://t.me/some_channel_with_keys",
            "mask":[
                "ss://"
            ],
            "configcount":3,
            "parsetoptobot":false 
        },
        {
            "url":"https://www.some_site_with_keys.com/",
            "mask":[
                "ss://"
            ],
            "configcount":1,
            "parsetoptobot":true 
        }
    ]
```
Links for parsing. In this section 
```
  "configcount" 
```
how many configs do you want to extract from this page
```
  "parsetoptobot"
```
if true parsing will done from top to bottom. 
- Use "true" for pages where new information placing at the top, like sites
- Use "false" for pages where new information placing at the bottom, like telegram channels and some forums
