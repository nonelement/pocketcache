### PocketCache

Exports your Pocketed items. Probably more like a dump than an export in reality.

##### config file

The config file expected, by default, is `pocketcache.config.json`

The config will look like the following:

```
{
    "APP_NAME": "Pocketcache",
    "CLIENT_KEY": "00000-000000000000000000000000",
    "ACCESS_TOKEN": "00000000-0000-0000-0000-000000",
    "REQUEST_TOKEN": "00000000-0000-0000-0000-000000"
}
```

Note: ACCESS_TOKEN, REQUEST_TOKEN can be blank -- those will be fetched by this app
through the authentication process.

##### export

Exported pocket entries adhere to Pocket's data model for the endpoints used. 
Output will be pretty printed, and file name will be `pocketcache.export.json` by default.  