# ssl-exporter

A simple prometheus exporter that returns the `NotBefore` and `NotAfter` properties of the primary certificate for a list of given domains.

## Usage

1. Clone repo
2. Build (`go build -o exporter main.go`)
3. Configure
    - See `config-example.json` for default values
    - Save changed version as `config.json` in working directory
4. Run (`./exporter`)
    - Or `./exporter -config ~/config.json` if the config is somewhere else.
5. (Optional) You can use the `example-systemd.service` file to create the service.
    - Make sure you edit the placeholder values to fit your setup!

### Queries

- You can view the validity of all your configured domains using a query like:
    - `cert_not_after - time()`
- To view the soonest expiring certificate (for example to use in a singlestat panel):
    - `bottomk(1, sum by (domain) (cert_not_after - time()))`

### Example response

```
# HELP cert_not_before The primary certificates NotBefore date as unix time'.
# TYPE cert_not_before gauge
cert_not_before{domain="example.com"} 1592149450
# HELP cert_not_after The primary certificates NotAfter date as unix time'.
# TYPE cert_not_after gauge
cert_not_after{domain="example.com"} 1599925450
# HELP cert_fetch_duration Duration of the http call in nanoseconds.
# TYPE cert_fetch_duration gauge
cert_fetch_duration{domain="example.com"} 482473832
# HELP cert_fetch_success Success of the http call as a 0/1 boolean.
# TYPE cert_fetch_success gauge
cert_fetch_success{domain="example.com"} 1
```
