package main

type modemChannel struct {
	channelID  int
	channel    int
	frequency  int
	snr        int
	power      int
	prerserr   int
	postrserr  int
	modulation string
	scheme     string

	noise       int
	attenuation int
}

type modemConfig struct {
	config   string
	maxrate  int
	maxburst int
}

type modemStats struct {
	configs      []modemConfig
	upChannels   []modemChannel
	downChannels []modemChannel
	fetchTime    int64
	modemType    string
}

type docsisModem interface {
	ParseStats() (modemStats, error)
	ClearStats()
	Type() string
}

var commandLineOpts struct {
	Daemon         bool `short:"d" long:"daemon" description:"Gather statistics on new line to STDIN?"`
	PrometheusPort int  `short:"p" long:"port" description:"Prometheus exporter port (disabled if not defined)"`
}
