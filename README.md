# Fediverse Matrix KeyGen

a self-service Matrix account creation and password reset utility for Mastodon or Pleroma users.

## How to Use
Make sure you have the following:
* a Mastodon or Pleroma server (e.g. mastodon.my.site) 
* a Matrix synapse server (e.g. matrix.my.site) 
* a Matrix user with admin privilige, see [synapse doc](https://github.com/matrix-org/synapse/blob/develop/docs/admin_api/README.rst) for how to grant admin privilige

Optionally you can have this service running behind a reverse proxy like nginx with some URL rule so that it looks like part of your existing service, e.g. it can become `https://mastodon.my.site/enter-the-matrix` by adding some configuration in nginx like the following:
```
upstream keygen {
    server          127.0.0.1:8848;
}

server {
    server_name     mastodon.my.site;
    ......
    location /enter-the-matrix {
        proxy_pass  http://keygen;
    }
}
```

Start the service with following parameters:
```
fediverse-matrix-keygen -m matrix.my.site -mu admin_user -mp admin_pass -f mastodon.my.site -u https://mastodon.my.site/enter-the-matrix -p 8848
```

Now users will be able to get their Matrix account from `https://mastodon.my.site/enter-the-matrix` .

## Notes
Depending on your Matrix configuration, sometimes server name and homeserver address might be different, make sure to use the server name (the part after `:` in `@id:matrix.org` ) with `-m` parameter.

The matrix password reset API will logout all sessions for that user, so the user is advised to always backup encryption keys and/or passphrases, otherwise they will lose access to all encrypted history chats.

## License
MIT