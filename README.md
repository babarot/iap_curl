iap_curl
========

curl wrapper for making HTTP request to IAP-protected app in CLI more easier than curl

## Usage

```console
$ export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
$ export IAP_CLIENT_ID="342624545358-asdfd8fas9df8sd7ga0sdguadfpvqp69.apps.googleusercontent.com"
$
$ iap_curl http://iap-protected.webapp.com
```

The option of iap_curl is fully compatible with curl one.

## Installation

```
$ go get github.com/b4b4r07/iap_curl
```

## License

MIT

## Author

b4b4r07
