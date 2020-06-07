# ssl-exporter

A simple prometheus exporter that returns the `NotAfter` property of given domains.

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
