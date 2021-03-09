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
}
