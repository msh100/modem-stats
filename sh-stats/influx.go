package main

import (
	"fmt"
	"strings"
)

func printForInflux(routerStats modemStats) {
	for _, downChannel := range routerStats.downChannels {
		var keys []string
		var values []string

		if routerStats.modemType == "VDSL" {
			keys = append(keys, fmt.Sprintf("id=%d", downChannel.channelID))
			values = append(
				values,
				fmt.Sprintf("noise=%d", downChannel.noise),
				fmt.Sprintf("attenuation=%d", downChannel.attenuation),
			)
		} else {
			keys = append(
				keys,
				fmt.Sprintf("channel=%d", downChannel.channel),
				fmt.Sprintf("id=%d", downChannel.channelID),
				fmt.Sprintf("modulation=%s", downChannel.modulation),
				fmt.Sprintf("scheme=%s", downChannel.scheme),
			)
			values = append(
				values,
				fmt.Sprintf("frequency=%d", downChannel.frequency),
				fmt.Sprintf("snr=%d", downChannel.snr),
				fmt.Sprintf("power=%d", downChannel.power),
				fmt.Sprintf("prerserr=%d", downChannel.prerserr),
				fmt.Sprintf("postrserr=%d", downChannel.postrserr),
			)
		}

		output := fmt.Sprintf(
			"downstream,%s %s",
			strings.Join(keys, ","),
			strings.Join(values, ","),
		)
		fmt.Println(output)
	}
	for _, upChannel := range routerStats.upChannels {
		var keys []string
		var values []string

		if routerStats.modemType == "VDSL" {
			keys = append(keys, fmt.Sprintf("id=%d", upChannel.channelID))
			values = append(
				values,
				fmt.Sprintf("noise=%d", upChannel.noise),
				fmt.Sprintf("attenuation=%d", upChannel.attenuation),
			)
		} else {
			keys = append(
				keys,
				fmt.Sprintf("channel=%d", upChannel.channel),
				fmt.Sprintf("id=%d", upChannel.channelID),
			)
			values = append(
				values,
				fmt.Sprintf("frequency=%d", upChannel.frequency),
				fmt.Sprintf("power=%d", upChannel.power),
			)
		}

		output := fmt.Sprintf(
			"upstream,%s %s",
			strings.Join(keys, ","),
			strings.Join(values, ","),
		)
		fmt.Println(output)
	}
	for _, config := range routerStats.configs {
		values := []string{
			fmt.Sprintf("maxrate=%d", config.maxrate),
		}
		if config.maxburst != 0 {
			values = append(values, fmt.Sprintf("maxburst=%d", config.maxburst))
		}

		output := fmt.Sprintf(
			"config,config=%s %s",
			config.config,
			strings.Join(values, ","),
		)
		fmt.Println(output)
	}

	fmt.Println(fmt.Sprintf("shstatsinfo timems=%d", routerStats.fetchTime))
}
