package main

import "fmt"

func printForInflux(routerStats routerStats) {
	for _, downChannel := range routerStats.downChannels {
		output := fmt.Sprintf(
			"downstream,channel=%d,id=%d,modulation=%s,scheme=%s frequency=%d,snr=%d,power=%d,prerserr=%d,postrserr=%d",
			downChannel.channel,
			downChannel.channelID,
			downChannel.modulation,
			downChannel.scheme,
			downChannel.frequency,
			downChannel.snr,
			downChannel.power,
			downChannel.prerserr,
			downChannel.postrserr,
		)
		fmt.Println(output)
	}
	for _, upChannel := range routerStats.upChannels {
		output := fmt.Sprintf(
			"upstream,channel=%d,id=%d frequency=%d,power=%d",
			upChannel.channel,
			upChannel.channelID,
			upChannel.frequency,
			upChannel.power,
		)
		fmt.Println(output)
	}
	for _, config := range routerStats.configs {
		output := fmt.Sprintf(
			"config,config=%s maxrate=%d,maxburst=%d",
			config.config,
			config.maxrate,
			config.maxburst,
		)
		fmt.Println(output)
	}

	fmt.Println(fmt.Sprintf("shstatsinfo timems=%d", routerStats.fetchTime))
}
