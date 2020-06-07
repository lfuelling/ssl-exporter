package main

import (
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Domains       []string
	ListenAddress string
}

var config Config

func loadCertificates() ([]*x509.Certificate, string) {
	startTime := time.Now().UnixNano()
	var certs []*x509.Certificate
	var fetchMetrics = ""

	for _, d := range config.Domains {
		log.Println("Loading certificate for '" + d + "'...")
		resp, err := http.Get("https://" + d + "/")
		if err != nil {
			log.Println("Error fetching '"+d+"'!", err)
			fetchMetrics += "cert_fetch_success{domain=\"" + d + "\"} 0\n"
		} else {
			if resp.TLS != nil {
				certificates := resp.TLS.PeerCertificates
				if len(certificates) > 0 {
					certs = append(certs, certificates[0])
					fetchMetrics += "cert_fetch_success{domain=\"" + d + "\"} 1\n"
				} else {
					log.Println("No certificates given for '" + d + "'!")
					fetchMetrics += "cert_fetch_success{domain=\"" + d + "\"} 0\n"
				}
			} else {
				log.Println("TLS properties nil for '"+d+"'!", err)
				fetchMetrics += "cert_fetch_success{domain=\"" + d + "\"} 0\n"
			}

		}
		fetchMetrics += "cert_fetch_duration{domain=\"" + d + "\"} " + strconv.FormatInt(time.Now().UnixNano()-startTime, 10) + "\n"
	}
	return certs, fetchMetrics
}

func renderMetricsResponse() (string, error) {
	certs, metrics := loadCertificates()

	res := "# HELP cert_not_before The primary certificates NotBefore date as unix time'.\n" +
		"# TYPE cert_not_before gauge\n" +
		"# HELP cert_not_after The primary certificates NotAfter date as unix time'.\n" +
		"# TYPE cert_not_after gauge\n" +
		"# HELP cert_fetch_duration Duration of the http call in nanoseconds.\n" +
		"# TYPE cert_fetch_duration gauge\n" +
		"# HELP cert_fetch_success Success of the http call as a 0/1 boolean.\n" +
		"# TYPE cert_fetch_success gauge\n" + metrics
	for _, crt := range certs {
		for _, dnsName := range crt.DNSNames {
			res += `cert_not_before{domain="` + dnsName + `"} ` + strconv.FormatInt(crt.NotBefore.Unix(), 10) + "\n"
			res += `cert_not_after{domain="` + dnsName + `"} ` + strconv.FormatInt(crt.NotAfter.Unix(), 10) + "\n"
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
		Addr:         config.ListenAddress,
		Handler:      http.HandlerFunc(handleMetrics),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("Starting server...")
	err2 := server.ListenAndServe()
	if err2 != nil {
		log.Fatalln("Unable to start server!", err2)
	}
}
