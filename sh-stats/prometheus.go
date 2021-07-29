package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusExporter struct {
	downFrequency   *prometheus.Desc
	downPower       *prometheus.Desc
	downSNR         *prometheus.Desc
	downPreRS       *prometheus.Desc
	downPostRS      *prometheus.Desc
	upFrequency     *prometheus.Desc
	upPower         *prometheus.Desc
	maxrate         *prometheus.Desc
	maxburst        *prometheus.Desc
	fetchtime       *prometheus.Desc
	downNoise       *prometheus.Desc
	downAttenuation *prometheus.Desc
	upNoise         *prometheus.Desc
	upAttenuation   *prometheus.Desc

	docsisModem docsisModem
}

func (p *PrometheusExporter) Collect(ch chan<- prometheus.Metric) {
	resetStats(p.docsisModem)
	modemStats, _ := fetchStats(p.docsisModem)

	for _, c := range modemStats.downChannels {
		var labels []string

		if modemStats.modemType == "VDSL" {
			labels = []string{
				strconv.Itoa(c.channelID),
			}

			ch <- prometheus.MustNewConstMetric(
				p.downNoise,
				prometheus.GaugeValue,
				float64(c.noise),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.downAttenuation,
				prometheus.GaugeValue,
				float64(c.attenuation),
				labels...,
			)
		} else {
			labels = []string{
				strconv.Itoa(c.channel),
				strconv.Itoa(c.channelID),
				c.modulation,
				c.scheme,
			}

			ch <- prometheus.MustNewConstMetric(
				p.downFrequency,
				prometheus.GaugeValue,
				float64(c.frequency),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.downPower,
				prometheus.GaugeValue,
				float64(c.power),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.downSNR,
				prometheus.GaugeValue,
				float64(c.snr),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.downPreRS,
				prometheus.GaugeValue,
				float64(c.prerserr),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.downPostRS,
				prometheus.GaugeValue,
				float64(c.postrserr),
				labels...,
			)
		}
	}

	for _, c := range modemStats.upChannels {
		var labels []string

		if modemStats.modemType == "VDSL" {
			labels = []string{
				strconv.Itoa(c.channelID),
			}

			ch <- prometheus.MustNewConstMetric(
				p.upNoise,
				prometheus.GaugeValue,
				float64(c.noise),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.upAttenuation,
				prometheus.GaugeValue,
				float64(c.attenuation),
				labels...,
			)
		} else {
			labels = []string{
				strconv.Itoa(c.channel),
				strconv.Itoa(c.channelID),
			}

			ch <- prometheus.MustNewConstMetric(
				p.upPower,
				prometheus.GaugeValue,
				float64(c.power),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.upFrequency,
				prometheus.GaugeValue,
				float64(c.frequency),
				labels...,
			)
		}
	}

	for _, config := range modemStats.configs {
		ch <- prometheus.MustNewConstMetric(
			p.maxrate,
			prometheus.GaugeValue,
			float64(config.maxrate),
			config.config,
		)
		if config.maxburst != 0 {
			ch <- prometheus.MustNewConstMetric(
				p.maxburst,
				prometheus.GaugeValue,
				float64(config.maxburst),
				config.config,
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(
		p.fetchtime,
		prometheus.GaugeValue,
		float64(modemStats.fetchTime),
	)
}

func (p *PrometheusExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.downFrequency
	ch <- p.upFrequency
	ch <- p.downPower
	ch <- p.upPower
	ch <- p.downSNR
	ch <- p.downPostRS
	ch <- p.downPreRS
	ch <- p.maxrate
	ch <- p.maxburst
	ch <- p.fetchtime
	ch <- p.downNoise
	ch <- p.downAttenuation
	ch <- p.upNoise
	ch <- p.upAttenuation
}

func ProExporter(docsisModem docsisModem) *PrometheusExporter {
	namespace := "modemstats"
	downLabels := []string{}
	upLabels := []string{}

	if docsisModem.Type() == "VDSL" {
		downLabels = []string{"id"}
		upLabels = []string{"id"}
	} else {
		downLabels = []string{"channel", "id", "modulation", "scheme"}
		upLabels = []string{"channel", "id"}
	}

	return &PrometheusExporter{
		docsisModem: docsisModem,
		downFrequency: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "downstream", "frequency"),
			"Downstream Frequency in HZ",
			downLabels,
			nil,
		),
		downPower: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "downstream", "power"),
			"Downstream Power level in dBmv",
			downLabels,
			nil,
		),
		downSNR: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "downstream", "snr"),
			"Downstream SNR in dB",
			downLabels,
			nil,
		),
		downPostRS: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "downstream", "postrserr"),
			"Number of Errors per channel Post RS",
			downLabels,
			nil,
		),
		downPreRS: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "downstream", "prerserr"),
			"Number of Errors per channel Pre RS",
			downLabels,
			nil,
		),
		downAttenuation: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "downstream", "attenuation"),
			"Downstream attenuation in TODO: wtf is this?",
			downLabels,
			nil,
		),
		downNoise: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "downstream", "noise"),
			"Downstream noise level in dB",
			downLabels,
			nil,
		),
		upFrequency: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "upstream", "frequency"),
			"Upstream Frequency in HZ",
			upLabels,
			nil,
		),
		upPower: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "upstream", "power"),
			"Upstream Power level in dBmv",
			upLabels,
			nil,
		),
		upAttenuation: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "upstream", "attenuation"),
			"Upstream attenuation in TODO: wtf is this?",
			downLabels,
			nil,
		),
		upNoise: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "upstream", "noise"),
			"Upstream noise level in dB",
			downLabels,
			nil,
		),
		maxrate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "config", "maxrate"),
			"Maximum link rate",
			[]string{"config"},
			nil,
		),
		maxburst: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "config", "maxburst"),
			"Maximum link burst rate",
			[]string{"config"},
			nil,
		),
		fetchtime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "shstatsinfo", "timems"),
			"Time to fetch statistics from the modem in milliseconds",
			[]string{},
			nil,
		),
	}
}

func modemStatsPrometheus(modem docsisModem, port int) {
	prometheus.Unregister(prometheus.NewGoCollector())
	exporter := ProExporter(modem)
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println(fmt.Sprintf("Starting Prometheus exporter on port %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
