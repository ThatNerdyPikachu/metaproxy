# metaproxy

## What is it?
This proxies requests to Plex's ``/library/metadata`` endpoints, and adds in subtitle and audio track information. The logic for actually changing this is based off [Ian Walton's userscript](https://forums.plex.tv/t/show-subtitle-audio-names-and-media-version-info-for-plex/552743) (Thanks Ian!)

## Setup
1. Download the [latest version of metaproxy](https://github.com/ThatNerdyPikachu/metaproxy/releases/latest) for your OS/arch, and unzip it somewhere.

2. Grab your Plex servers ``*.<hash>.plex.direct`` certificate
- That's inside the main app data directory on Windows & Linux ([see this Plex support article for more information](https://support.plex.tv/articles/202915258-where-is-the-plex-media-server-data-directory-located/))
- On macOS, it's in the regular library caches (``~/Library/Caches/PlexMediaServer/certificate.p12``)
- Thanks to Plex for providing me with this information!

3. Decrypt your certificate
- The key to this is equal to the SHA512 **hash** of the string ``plex<MACHINE ID>``
    - Where ``<MACHINE ID>`` is equal to the value of ``ProcessedMachineIdentifier`` from your servers advanced settings, [see this Plex support article for more information](https://support.plex.tv/articles/201105343-advanced-hidden-server-settings/)
- So, if my machine ID was ``deadbeef``, the string to hash would be ``plexdeadbeef``.
- The key is lowercase, as is the string to hash.
Example commands to decrypt the certificate:
```
openssl pkcs12 -in certificate.p12 -out plex.cert -clcerts -nokeys -passin "pass:<that hash>"
openssl pkcs12 -in certificate.p12 -out plex.key -nocerts -nodes -passin "pass:<that hash>"
```
where ``<that hash>`` is your hash.

4. Configure your Plex server
Currently I've only been able to get this working on Docker, as with that, I can map a port on the container to a port on my server.

My container is ran as so:
```
docker run -e ADVERTISE_IP="http://<local ip>:32400/" -e ALLOWED_NETWORKS=192.168.200.0/24 \
-d --restart=always \
--name plex \
-p 32401:32400 \
-e TZ="<timezone>" \
-v <redacted>:/config \
-v <redacted>:/media:ro \
plexinc/pms-docker
```
(If there is something I am configuring wrong here, please open an issue, PR, or post on the [Plex forum thread](https://forums.plex.tv/t/metaproxy-for-plex/566250). Thanks!)

5. Configure your reverse proxy
- I use [Caddy](https://caddyserver.com/v1), but feel free to use what you want!

My configuration is as so (note the ``*``):
```
*.<subdomain>.plex.direct:32400 {
        proxy /library/metadata localhost:3213 {
                transparent
        }

        proxy / localhost:32401 {
                transparent
                websocket
        }

        log logs/plex.log

        tls plex.cert plex.key
}
```
where ``<subdomain>`` is equal to the value found at the top of the certificate you decrypted. Example:
```
subject=/C=US/ST=California/L=Los Gatos/O=Plex, Inc./CN=*.<subdomain>.plex.direct
```

6. Configure metaproxy
Congrats! We're at the last step. Configuration options are as follows (you can also see this by running ``metaproxy -h``):
```
Usage of ./metaproxy:
  -addr string
        the address to bind to (default "127.0.0.1:3213")
  -plex-host string
        the host + port that your plex server is running on (default "localhost:32401")
  -secure
        use https to connect to your plex server (will increase loading times) (needed if Secure Connections is set to Required)
```

The default options work fine for the setup that we've configured in this guide, but make sure to change anything accordingly!
**Important:** If your server is set to require secure connections, you need to pass ``-secure``.

Congrats, you're done!

## Support
If you run into any issues, please post in the [Plex forum thread](https://forums.plex.tv/t/metaproxy-for-plex/566250), or open an issue here.