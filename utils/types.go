package utils

type ModemChannel struct {
	ChannelID  int
	Channel    int
	Frequency  int
	Snr        int
	Power      int
	Prerserr   int
	Postrserr  int
	Modulation string
	Scheme     string

	Noise       int
	Attenuation int
}

type ModemConfig struct {
	Config   string
	Maxrate  int
	Maxburst int
}

type ModemStats struct {
	Configs      []ModemConfig
	UpChannels   []ModemChannel
	DownChannels []ModemChannel
	FetchTime    int64
	ModemType    string
}

type DocsisModem interface {
	ParseStats() (ModemStats, error)
	ClearStats()
	Type() string
}

const (
	TypeDocsis = "DOCSIS"
	TypeVDSL   = "VDSL"
)
