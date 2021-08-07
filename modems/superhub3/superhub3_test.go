package superhub3

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/msh100/modem-stats/utils"
	"github.com/stretchr/testify/assert"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func Test_ClearStats(t *testing.T) {
	modemWithStats := Modem{
		Stats: []byte("random"),
	}

	assert.Equal(t, []byte("random"), modemWithStats.Stats)
	modemWithStats.ClearStats()
	assert.Equal(t, []byte(nil), modemWithStats.Stats)
}

func Test_Type(t *testing.T) {
	modem := Modem{}
	assert.Equal(t, "DOCSIS", modem.Type())
}
func Test_fetchURL(t *testing.T) {
	definedIP := Modem{
		IPAddress: "192.168.0.1",
	}
	defaultIP := Modem{}

	assert.Equal(t, "http://192.168.0.1/getRouterStatus", definedIP.fetchURL())
	assert.Equal(t, "http://192.168.100.1/getRouterStatus", defaultIP.fetchURL())
}

func Test_activeChannels(t *testing.T) {
	type test struct {
		input string
		up    []int
	}

	expectedDown := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}
	expectedUp := []int{1, 2, 3, 4}
	testTable := []test{
		{
			input: "chm.json",
		},
		{
			input: "chmb.json",
		},
		{
			input: "chmi.json",
		},
		{
			input: "eni.json",
		},
		{
			input: "frd.json",
		},
		{
			input: "msh.json",
		},
		{
			input: "mshm.json",
			up:    []int{2, 3, 4, 5}, // This user has channel 1 used for their telephone over DOCSIS
		},
	}

	for _, testData := range testTable {
		dat, err := ioutil.ReadFile(fmt.Sprintf("test_state/%s", testData.input))
		check(err)
		fmt.Println(testData.input)
		modem := Modem{
			Stats: dat,
		}

		testUp := expectedUp
		if testData.up != nil {
			testUp = testData.up
		}

		downChannels, upChannels := modem.activeChannels()
		assert.EqualValues(t, expectedDown, downChannels)
		assert.EqualValues(t, testUp, upChannels)
	}
}

func Test_activeProfile(t *testing.T) {
	type test struct {
		input       string
		upProfile   int
		downProfile int
	}

	testTable := []test{
		{
			input:       "chm.json",
			upProfile:   137484,
			downProfile: 137483,
		},
		{
			input:       "chmb.json",
			upProfile:   8805,
			downProfile: 8804,
		},
		{
			input:       "chmi.json",
			upProfile:   8805,
			downProfile: 8804,
		},
		{
			input:       "eni.json",
			upProfile:   981273,
			downProfile: 981272,
		},
		{
			input:       "frd.json",
			upProfile:   172472,
			downProfile: 172471,
		},
		{
			input:       "msh.json",
			upProfile:   12500,
			downProfile: 12499,
		},
		{
			input:       "mshm.json",
			upProfile:   12428,
			downProfile: 12427,
		},
	}

	for _, testData := range testTable {
		dat, err := ioutil.ReadFile(fmt.Sprintf("test_state/%s", testData.input))
		check(err)
		modem := Modem{
			Stats: dat,
		}

		downProfile, upProfile := modem.activeConfigs()
		assert.EqualValues(t, testData.upProfile, downProfile, fmt.Sprintf("%s up profile mismatch", testData.input))
		assert.EqualValues(t, testData.downProfile, upProfile, fmt.Sprintf("%s down profile mismatch", testData.input))
	}
}

func Test_readMIBInt(t *testing.T) {
	type testValues struct {
		mib      string
		finalInt int
		expected int
	}
	type test struct {
		input string
		tests []testValues
	}

	testTable := []test{
		{
			input: "chm.json",
			tests: []testValues{
				{
					mib:      "1.3.6.1.4.1.4491.2.1.21.1.2.1.7.2.3",
					finalInt: 137534,
					expected: 3044,
				},
				{
					mib:      "1.3.6.1.4.1.4491.2.1.20.1.2.1.2",
					finalInt: 4,
					expected: 36,
				},
				{
					mib:      "1337",
					finalInt: 1234,
					expected: -1, // Intentionally invalid
				},
			},
		},
		{
			input: "msh.json",
			tests: []testValues{
				{
					mib:      "1.3.6.1.4.1.4491.2.1.20.1.2.1.2",
					finalInt: 4,
					expected: 5,
				},
			},
		},
	}

	for _, testData := range testTable {
		dat, err := ioutil.ReadFile(fmt.Sprintf("test_state/%s", testData.input))
		check(err)
		modem := Modem{
			Stats: dat,
		}
		dataJSON := modem.dataAsJSON()

		for _, testCase := range testData.tests {
			MibValue := modem.readMIBInt(dataJSON, testCase.mib, testCase.finalInt)
			assert.EqualValues(t, testCase.expected, MibValue)
		}

	}
}

func Test_dataAsJSON(t *testing.T) {
	testTable := []string{
		"chm.json",
		"chmb.json",
		"chmi.json",
		"eni.json",
		"frd.json",
		"msh.json",
		"mshm.json",
	}
	mibRegex := "^([0-2])((.0)|(.[1-9][0-9]*))*$"
	re := regexp.MustCompile(mibRegex)

	for _, testData := range testTable {
		dat, err := ioutil.ReadFile(fmt.Sprintf("test_state/%s", testData))
		check(err)
		modem := Modem{
			Stats: dat,
		}
		dataJSON := modem.dataAsJSON()

		for key := range dataJSON {
			assert.True(t, re.Match([]byte(key)))
		}
	}
}

func Test_ParseStats(t *testing.T) {
	testTable := []string{
		"chm.json",
		"chmb.json",
		"chmi.json",
		"eni.json",
		"frd.json",
		"msh.json",
		"mshm.json",
	}

	for _, testData := range testTable {
		dat, err := ioutil.ReadFile(fmt.Sprintf("test_state/%s", testData))
		check(err)
		modem := Modem{
			Stats: dat,
		}

		parsed, err := modem.ParseStats()
		assert.Nil(t, err)
		assert.Equal(t, "DOCSIS", parsed.ModemType)
		assert.Zero(t, parsed.FetchTime)

		assert.Greater(t, len(parsed.DownChannels), 0)
		for _, downChannel := range parsed.DownChannels {
			validateDownChannel(t, downChannel)
		}

		assert.Greater(t, len(parsed.UpChannels), 0)
		for _, upChannel := range parsed.UpChannels {
			validateUpChannel(t, upChannel)
		}

		assert.Equal(t, len(parsed.Configs), 2)
		if parsed.Configs[0].Config == "upstream" {
			assert.Equal(t, "upstream", parsed.Configs[0].Config)
			assert.Equal(t, "downstream", parsed.Configs[1].Config)
		} else {
			assert.Equal(t, "downstream", parsed.Configs[0].Config)
			assert.Equal(t, "upstream", parsed.Configs[1].Config)
		}
		assert.Greater(t, parsed.Configs[0].Maxrate, 0)
		assert.Greater(t, parsed.Configs[1].Maxrate, 0)
		assert.Greater(t, parsed.Configs[0].Maxburst, 0)
		assert.Greater(t, parsed.Configs[1].Maxburst, 0)
	}

}

func validateDownChannel(t *testing.T, channel utils.ModemChannel) {
	assert.Greater(t, channel.ChannelID, 0)
	assert.Greater(t, channel.Channel, 0)

	assert.Greater(t, channel.Frequency, 108000000) // DOCSIS 3.0 108Mhz to 1Ghz
	assert.Less(t, channel.Frequency, 1000000000)

	assert.Greater(t, channel.Snr, 30) // DOCSIS 3.0 spec has 30dB minimum

	assert.Greater(t, channel.Power, -150) // DOCSIS 3.0 spec allows +/- 15dBmV
	assert.Less(t, channel.Power, 150)

	assert.GreaterOrEqual(t, channel.Prerserr, 0)
	assert.GreaterOrEqual(t, channel.Postrserr, 0)

	assert.Equal(t, "QAM256", channel.Modulation)
	assert.Equal(t, "SC-QAM", channel.Scheme)

	assert.Zero(t, channel.Noise)
	assert.Zero(t, channel.Attenuation)
}

func validateUpChannel(t *testing.T, channel utils.ModemChannel) {
	assert.Greater(t, channel.ChannelID, 0)
	assert.Greater(t, channel.Channel, 0)

	assert.Greater(t, channel.Frequency, 5000000) // DOCSIS 3.0 5Mhz to 85Mhz
	assert.Less(t, channel.Frequency, 85000000)

	assert.NotZero(t, channel.Power)

	assert.Equal(t, channel.Prerserr, 0)
	assert.Equal(t, channel.Postrserr, 0)

	assert.Equal(t, "", channel.Modulation)
	assert.Equal(t, "", channel.Scheme)

	assert.Zero(t, channel.Noise)
	assert.Zero(t, channel.Attenuation)
}

func Test_FailedHTTPCall(t *testing.T) {
	modem := Modem{
		IPAddress: "127.0.0.1",
	}

	stats, err := modem.ParseStats()
	assert.NotNil(t, err)
	assert.Equal(t, stats, utils.ModemStats{})

	modem = Modem{
		IPAddress: "127.0.0.257",
	}

	stats, err = modem.ParseStats()
	assert.NotNil(t, err)
	assert.Equal(t, stats, utils.ModemStats{})
}
