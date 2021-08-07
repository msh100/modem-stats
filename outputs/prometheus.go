package outputs

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/msh100/modem-stats/utils"
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

	docsisModem utils.DocsisModem
}

func (p *PrometheusExporter) Collect(ch chan<- prometheus.Metric) {
	utils.ResetStats(p.docsisModem)
	modemStats, _ := utils.FetchStats(p.docsisModem)

	for _, c := range modemStats.DownChannels {
		var labels []string

		if modemStats.ModemType == utils.TypeVDSL {
			labels = []string{
				strconv.Itoa(c.ChannelID),
			}

			ch <- prometheus.MustNewConstMetric(
				p.downNoise,
				prometheus.GaugeValue,
				float64(c.Noise),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.downAttenuation,
				prometheus.GaugeValue,
				float64(c.Attenuation),
				labels...,
			)
		} else {
			labels = []string{
				strconv.Itoa(c.Channel),
				strconv.Itoa(c.ChannelID),
				c.Modulation,
				c.Scheme,
			}

			ch <- prometheus.MustNewConstMetric(
				p.downFrequency,
				prometheus.GaugeValue,
				float64(c.Frequency),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.downPower,
				prometheus.GaugeValue,
				float64(c.Power),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.downSNR,
				prometheus.GaugeValue,
				float64(c.Snr),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.downPreRS,
				prometheus.GaugeValue,
				float64(c.Prerserr),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.downPostRS,
				prometheus.GaugeValue,
				float64(c.Postrserr),
				labels...,
			)
		}
	}

	for _, c := range modemStats.UpChannels {
		var labels []string

		if modemStats.ModemType == utils.TypeVDSL {
			labels = []string{
				strconv.Itoa(c.ChannelID),
			}

			ch <- prometheus.MustNewConstMetric(
				p.upNoise,
				prometheus.GaugeValue,
				float64(c.Noise),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.upAttenuation,
				prometheus.GaugeValue,
				float64(c.Attenuation),
				labels...,
			)
		} else {
			labels = []string{
				strconv.Itoa(c.Channel),
				strconv.Itoa(c.ChannelID),
			}

			ch <- prometheus.MustNewConstMetric(
				p.upPower,
				prometheus.GaugeValue,
				float64(c.Power),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				p.upFrequency,
				prometheus.GaugeValue,
				float64(c.Frequency),
				labels...,
			)
		}
	}

	for _, config := range modemStats.Configs {
		ch <- prometheus.MustNewConstMetric(
			p.maxrate,
			prometheus.GaugeValue,
			float64(config.Maxrate),
			config.Config,
		)
		if config.Maxburst != 0 {
			ch <- prometheus.MustNewConstMetric(
				p.maxburst,
				prometheus.GaugeValue,
				float64(config.Maxburst),
				config.Config,
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(
		p.fetchtime,
		prometheus.GaugeValue,
		float64(modemStats.FetchTime),
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

func ProExporter(docsisModem utils.DocsisModem) *PrometheusExporter {
	namespace := "modemstats"
	downLabels := []string{}
	upLabels := []string{}

	if docsisModem.Type() == utils.TypeVDSL {
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

func Prometheus(modem utils.DocsisModem, port int) {
	prometheus.Unregister(prometheus.NewGoCollector())
	exporter := ProExporter(modem)
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println(fmt.Sprintf("Starting Prometheus exporter on port %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
