package outputs

import (
	"fmt"
	"strings"

	"github.com/msh100/modem-stats/utils"
)

func PrintForInflux(routerStats utils.ModemStats) {
	for _, downChannel := range routerStats.DownChannels {
		var keys []string
		var values []string

		if routerStats.ModemType == utils.TypeVDSL {
			keys = append(keys, fmt.Sprintf("id=%d", downChannel.ChannelID))
			values = append(
				values,
				fmt.Sprintf("noise=%d", downChannel.Noise),
				fmt.Sprintf("attenuation=%d", downChannel.Attenuation),
			)
		} else {
			keys = append(
				keys,
				fmt.Sprintf("channel=%d", downChannel.Channel),
				fmt.Sprintf("id=%d", downChannel.ChannelID),
				fmt.Sprintf("modulation=%s", downChannel.Modulation),
				fmt.Sprintf("scheme=%s", downChannel.Scheme),
			)
			values = append(
				values,
				fmt.Sprintf("frequency=%d", downChannel.Frequency),
				fmt.Sprintf("snr=%d", downChannel.Snr),
				fmt.Sprintf("power=%d", downChannel.Power),
				fmt.Sprintf("prerserr=%d", downChannel.Prerserr),
				fmt.Sprintf("postrserr=%d", downChannel.Postrserr),
			)
		}

		output := fmt.Sprintf(
			"downstream,%s %s",
			strings.Join(keys, ","),
			strings.Join(values, ","),
		)
		fmt.Println(output)
	}
	for _, upChannel := range routerStats.UpChannels {
		var keys []string
		var values []string

		if routerStats.ModemType == utils.TypeVDSL {
			keys = append(keys, fmt.Sprintf("id=%d", upChannel.ChannelID))
			values = append(
				values,
				fmt.Sprintf("noise=%d", upChannel.Noise),
				fmt.Sprintf("attenuation=%d", upChannel.Attenuation),
			)
		} else {
			keys = append(
				keys,
				fmt.Sprintf("channel=%d", upChannel.Channel),
				fmt.Sprintf("id=%d", upChannel.ChannelID),
			)
			values = append(
				values,
				fmt.Sprintf("frequency=%d", upChannel.Frequency),
				fmt.Sprintf("power=%d", upChannel.Power),
			)
		}

		output := fmt.Sprintf(
			"upstream,%s %s",
			strings.Join(keys, ","),
			strings.Join(values, ","),
		)
		fmt.Println(output)
	}
	for _, config := range routerStats.Configs {
		values := []string{
			fmt.Sprintf("maxrate=%d", config.Maxrate),
		}
		if config.Maxburst != 0 {
			values = append(values, fmt.Sprintf("maxburst=%d", config.Maxburst))
		}

		output := fmt.Sprintf(
			"config,config=%s %s",
			config.Config,
			strings.Join(values, ","),
		)
		fmt.Println(output)
	}

	fmt.Println(fmt.Sprintf("shstatsinfo timems=%d", routerStats.FetchTime))
}
