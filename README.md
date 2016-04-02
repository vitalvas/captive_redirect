# Captive redirect

## ENV Config

* `CAPTIVE_5555` - Where `5555` - port. Value - redirect path.
* `CAPTIVE_NOORIGIN` - ~~see code :)~~

## Usage

### Build and run

```
GOOS=linux GOARCH=amd64 go build -o /usr/local/sbin/captive_redirect main.go
```

Add to rc.local:

```
CAPTIVE_5555=http://my.example.com/no_money.cgi CAPTIVE_5556=http://acs.example.com/register.cgi /usr/local/sbin/captive_redirect &
```

### IPTables
```
iptables -t nat -A PREROUTING -m set --match-set no_money src -p tcp --dport 80 -j REDIRECT --to-port 5555
iptables -t nat -A PREROUTING -m set --match-set guest src -p tcp --dport 80 -j REDIRECT --to-port 5556
```
