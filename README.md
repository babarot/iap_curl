iap_curl
========

iap_curl is a curl wrapper for making HTTP request to IAP-protected app in CLI more easier than curl

## Usage

```console
$ export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
$ export IAP_CLIENT_ID="342624545358-asdfd8fas9df8sd7ga0sdguadfpvqp69.apps.googleusercontent.com"
$ iap_curl http://iap-protected.webapp.com
```

The options of `iap_curl` are fully compatible with curl.

If you want to use [httpstat](https://github.com/b4b4r07/httpstat), please specify the `IAP_CURL_BIN` environment variable:

```console
$ export IAP_CURL_BIN="httpstat.sh"
$ iap_curl https://tellme.tokyo
Connected to 104.31.70.103:443

HTTP/2.0 200 OK
Server: cloudflare-nginx
Access-Control-Allow-Origin: *
Cache-Control: max-age=600
Cf-Ray: 3af48c40aa3694cf-NRT
Content-Type: text/html; charset=utf-8
Date: Tue, 17 Oct 2017 16:13:54 GMT
Expires: Tue, 17 Oct 2017 16:23:54 GMT
Last-Modified: Mon, 16 Oct 2017 04:33:46 GMT
Set-Cookie: __cfduid=db7e1d73f138bcb26e0d6a040e9f5df491508256834; expires=Wed, 17-Oct-18 16:13:54 GMT; path=/; domain=.tellme.tokyo; HttpOnly; Secure
Strict-Transport-Security: max-age=15552000; preload
X-Content-Type-Options: nosniff
X-Github-Request-Id: 2A8B:16E6:10351CA:186074E:59E62C3F

Body discarded

  DNS Lookup   TCP Connection   TLS Handshake   Server Processing   Content Transfer
[      2ms  |          57ms  |        320ms  |            303ms  |             0ms  ]
            |                |               |                   |                  |
   namelookup:2ms            |               |                   |                  |
                       connect:60ms          |                   |                  |
                                   pretransfer:381ms             |                  |
                                                     starttransfer:684ms            |
                                                                                total:684ms
```

## Advanced usage

You can save the URL of frequently used service together with its Env (`IAP_CLIENT_ID` ...) in a JSON file (see also [#1](https://github.com/b4b4r07/iap_curl/issues/1)). This file is located in `~/.config/iap_curl/config.json`.

```json
{
  "services": [
    {
      "url": "https://my.service.com/health",
      "env": {
        "GOOGLE_APPLICATION_CREDENTIALS": "/Users/b4b4r07/Downloads/my-service-dev-b5e624fd28ee.json",
        "IAP_CLIENT_ID": "839558305167-s3akt4doo38lckhaac1ucfdp0e4921tc.apps.googleusercontent.com",
        "IAP_CURL_BIN": "curl"
      }
    }
  ]
}
```

Thanks to that, you can access more easier like curl.

```console
$ iap_curl https://my.service.com/health
```

Also, some original options are added. So you can use more and more easier to access the service by using [peco](https://github.com/peco/peco)/[fzf](https://github.com/junegunn/fzf). For more information about its options, please see `iap_curl --help`.

```console
$ iap_curl $(iap_curl --list-urls | peco) # peco is similar to fzf
```

## Installation

```
$ go get github.com/b4b4r07/iap_curl
```

## License

MIT

## Author

b4b4r07
