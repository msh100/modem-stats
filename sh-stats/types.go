package main

type downChannel struct {
	channelID  int
	channel    int
	frequency  int
	snr        int
	power      int
	prerserr   int
	postrserr  int
	modulation string
}
type upChannel struct {
	channelID int
	channel   int
	frequency int
	power     int
}
type config struct {
	config   string
	maxrate  int
	maxburst int
}

type routerStats struct {
	configs      []config
	upChannels   []upChannel
	downChannels []downChannel
}
