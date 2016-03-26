# Captive redirect

## ENV Config

* `CAPTIVE_5555` - Where `5555` - port. Value - redirect path.
* `CAPTIVE_NOORIGIN` - ~~see code :)~~

## Usage

### IPTables
```
iptables -t nat -A PREROUTING -m set --match-set guest src -p tcp --dport 80 -j REDIRECT --to-port 5555
```
