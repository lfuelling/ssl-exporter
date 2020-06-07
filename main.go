package main

import (
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Domains []string
	Port    int
}

var config Config

func loadCertificates() ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	for _, d := range config.Domains {
		log.Println("Loading certificate for '" + d + "'...")
		resp, err := http.Get("https://" + d + "/")
		if err != nil {
			log.Fatalln("Error fetching '"+d+"'!", err)
		} else {
			if resp.TLS != nil {
				certificates := resp.TLS.PeerCertificates
				if len(certificates) > 0 {
					certs = append(certs, certificates[0])
				} else {
					log.Fatalln("No certificates given for '" + d + "'!")
				}
			} else {
				log.Fatalln("TLS properties nil for '"+d+"'!", err)
			}

		}
	}
	return certs, nil
}

func renderMetricsResponse() (string, error) {
	certs, err := loadCertificates()
	if err != nil {
		return "", err
	}

	res := "# HELP cert_not_after The primary certificate's NotAfter date as unix time'.\n" +
		"# TYPE cert_not_after gauge"
	for _, crt := range certs {
		for _, dnsName := range crt.DNSNames {
			res += `cert_not_after{domain="` + dnsName + `"} ` + string(crt.NotAfter.Unix())
		}
	}
	return res, nil
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/metrics" {
		response, err := renderMetricsResponse()
		if err != nil {
			log.Println("Error fetching metrics!", err)
			w.WriteHeader(500)
			return
		}
		_, _ = fmt.Fprint(w, response)
	} else {
		log.Println("Not found: '" + r.RequestURI)
		w.WriteHeader(404)
	}
}

func main() {
	var configPath = flag.String("config", "config.json", "path to the config file")
	flag.Parse()
	file, err := os.Open(*configPath)
	defer file.Close()
	if err != nil {
		log.Fatalln("Unable to open config file!", err)
		return
	}
	decoder := json.NewDecoder(file)
	config = Config{}
	err1 := decoder.Decode(&config)
	if err1 != nil {
		log.Fatalln("Unable to read config!", err1)
		return
	}

	server := &http.Server{
		Addr:         "localhost:" + string(config.Port),
		Handler:      http.HandlerFunc(handleMetrics),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("Starting server...")
	log.Fatalln("Unable to start server!", server.ListenAndServe())
}
