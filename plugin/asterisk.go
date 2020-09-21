package asterisk

import (
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Asterisk struct {
	AsteriskIP  string `toml:"asterisk_ip"`
	AmiPort     int    `toml:"ami_port"`
	AmiUser     string `toml:"ami_user"`
	AmiPassword string `toml:"ami_password"`
}

var AsteriskConfig = `
  ##Sample Config
  #asterisk_ip = "0.0.0.0"
  #ami_port = 5038
  #ami_user = "user"
  #ami_password = "password"
`

func command(command string, s *Asterisk) string {
	var login = "Action: Login\r\nEvents: off\r\nUsername: " + s.AmiUser + "\r\nSecret: " + s.AmiPassword + "\r\n\r\n"
	var logoff = "Action: Logoff\r\n"

	data := ""
	finalCommand := login

	finalCommand = finalCommand + "Action: Command\r\nCommand: " + command + "\r\n\r\n"

	finalCommand = finalCommand + logoff

	data = SocketClient(s.AsteriskIP, s.AmiPort, finalCommand)
	_ = data
	return data
}

func SocketClient(ip string, port int, message string) string {

	socketResponse := ""
	const StopCharacter = "\r\n\r\n"
	messageInBytes := []byte(message)
	addr := strings.Join([]string{ip, strconv.Itoa(port)}, ":")
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		log.Fatalln(err)
	} else {

		defer conn.Close()

		conn.Write([]byte(messageInBytes))
		conn.Write([]byte(StopCharacter))

		var response, _ = ioutil.ReadAll(conn)
		socketResponse = string(response)
	}

	return socketResponse
}

func (s *Asterisk) SampleConfig() string {
	return AsteriskConfig
}

func (s *Asterisk) Description() string {
	return "Collects basic metrics dfrom Asterisk Open Source VoIP Software"
}

func (s *Asterisk) Gather(acc telegraf.Accumulator) error {

	//Call Volume
	callVolume := command("core show calls", s)
	callVolume = strings.Replace(callVolume, "\r\n", "\n", -1)
	currentCallVolume := strings.Split(callVolume, "\n")

	procesedCallVolume := strings.Replace(currentCallVolume[7], " calls processed", "", -1)
	procesedCallVolume = strings.TrimSpace(procesedCallVolume)
	currentCallVol := strings.Replace(currentCallVolume[6], "active call", "", -1)
	currentCallVol = strings.Replace(currentCallVol, "s", "", -1)
	currentCallVol = strings.TrimSpace(currentCallVol)

	currentCallValue, _ := strconv.Atoi(currentCallVol)

	//PRI Channels
	priResult := command("pri show channels", s)
	priResult = strings.Replace(priResult, "\r\n", "\n", -1)
	priResults := strings.Split(priResult, "\n")

	priOpenchannels := 0

	if !strings.Contains(priResult, "No such command") {

		priOpenchannels := 0
		for _, priItem := range priResults {

			//PRI B Chan Call PRI Channel Span Chan Chan Idle Level Call Name
			//1 1 Yes No Idle Yes
			if strings.Contains(priItem, "No") {
				priOpenchannels = priOpenchannels + 1
			}
		}
	}

	//SIP Peers
	sipResult := command("sip show peers", s)
	sipResult = strings.Replace(sipResult, "\r\n", "\n", -1)

	sipPeersCount := 0
	onlineMonitoredPeers := 0
	offlineMonitoredPeers := 0
	onlineUnmonitoredPeers := 0
	offlineUnmonitoredPeers := 0

	if !strings.Contains(sipResult, "No such command") {

		re := regexp.MustCompile(`([0-9]+) sip peer`)
		sipPeers := re.FindAllStringSubmatch(sipResult, -1)

		re = regexp.MustCompile(`Monitored: ([0-9]+) online, ([0-9]+) offline`)
		monitoredPeers := re.FindAllStringSubmatch(sipResult, -1)

		re = regexp.MustCompile(`Unmonitored: ([0-9]+) online, ([0-9]+) offline`)
		unMonitoredPeers := re.FindAllStringSubmatch(sipResult, -1)

		sipPeersCount, _ = strconv.Atoi(sipPeers[0][1])
		onlineMonitoredPeers, _ = strconv.Atoi(monitoredPeers[0][1])
		offlineMonitoredPeers, _ = strconv.Atoi(monitoredPeers[0][2])
		onlineUnmonitoredPeers, _ = strconv.Atoi(unMonitoredPeers[0][1])
		offlineUnmonitoredPeers, _ = strconv.Atoi(unMonitoredPeers[0][2])
	}

	//IAX2 Peers
	iaxResult := command("iax2 show peers", s)
	iaxResult = strings.Replace(iaxResult, "\r\n", "\n", -1)

	iaxPeersTotalCount := 0
	iaxPeersOnlineCount := 0
	iaxPeersOfflineCount := 0
	iaxPeersUnmonitoredCount := 0

	if !strings.Contains(iaxResult, "No such command") {

		re := regexp.MustCompile(`([0-9]+) iax2 peers`)
		iaxPeersTotal := re.FindAllStringSubmatch(iaxResult, -1)

		re = regexp.MustCompile(`\[([0-9]+) online`)
		iaxPeersOnline := re.FindAllStringSubmatch(iaxResult, -1)

		re = regexp.MustCompile(`([0-9]+) offline`)
		iaxPeersOffline := re.FindAllStringSubmatch(iaxResult, -1)

		re = regexp.MustCompile(`([0-9]+) unmonitored`)
		iaxPeersUnmonitored := re.FindAllStringSubmatch(iaxResult, -1)

		iaxPeersTotalCount, _ = strconv.Atoi(iaxPeersTotal[0][1])
		iaxPeersOnlineCount, _ = strconv.Atoi(iaxPeersOnline[0][1])
		iaxPeersOfflineCount, _ = strconv.Atoi(iaxPeersOffline[0][1])
		iaxPeersUnmonitoredCount, _ = strconv.Atoi(iaxPeersUnmonitored[0][1])
	}

	//DAHDI Trunks
	dahdiResult := command("dahdi show status", s)

	dahdiResult = strings.Replace(dahdiResult, "\r\n", "\n", -1)
	dahdiResults := strings.Split(dahdiResult, "\n")

	dahdiTotalTrunks := 0
	dahdiOnlineTrunks := 0
	dahdiOfflineTrunks := 0

	if !strings.Contains(iaxResult, "No such command") {

		for _, dahdiItem := range dahdiResults {
			if strings.Contains(dahdiItem, "Wildcard") || strings.Contains(dahdiItem, "wanpipe") {
				if strings.Contains(dahdiItem, "OK") {
					dahdiTotalTrunks = dahdiTotalTrunks + 1
					dahdiOnlineTrunks = dahdiOnlineTrunks + 1
				}
				if strings.Contains(dahdiItem, "RED") || strings.Contains(dahdiItem, "YEL") || strings.Contains(dahdiItem, "UNCONFI") {
					dahdiTotalTrunks = dahdiTotalTrunks + 1
					dahdiOfflineTrunks = dahdiOfflineTrunks + 1
				}
			}
		}

	}

	//G729 Codecs
	g729Result := command("g729 show licenses", s)
	g729Result = strings.Replace(g729Result, "\r\n", "\n", -1)

	g729TotalCount := 0
	g729EncodersCount := 0
	g729DecodersCount := 0

	if !strings.Contains(g729Result, "No such command") {

		re := regexp.MustCompile(`([0-9]+) licensed`)
		g729Total := re.FindAllStringSubmatch(g729Result, -1)

		re = regexp.MustCompile(`([0-9]+) encoders/decoders`)
		g729Encoders := re.FindAllStringSubmatch(g729Result, -1)

		re = regexp.MustCompile(`([0-9]+) encoders/decoders`)
		g729Decoders := re.FindAllStringSubmatch(g729Result, -1)

		g729TotalCount, _ = strconv.Atoi(g729Total[0][1])
		g729EncodersCount, _ = strconv.Atoi(g729Encoders[0][1])
		g729DecodersCount, _ = strconv.Atoi(g729Decoders[0][1])

	}

	//Asterisk Uptime
	uptimeResult := command("core show uptime", s)
	uptimeResult = strings.Replace(uptimeResult, "\r\n", "\n", -1)
	uptimeResults := strings.Split(uptimeResult, "\n")

	systemUptime := 0
	systemUptimeYears := 0
	systemUptimeWeeks := 0
	systemUptimeDays := 0
	systemUptimeHours := 0
	systemUptimeMinutes := 0
	systemUptimeSeconds := 0

	asteriskLastReload := 0
	asteriskLastReloadYears := 0
	asteriskLastReloadWeeks := 0
	asteriskLastReloadDays := 0
	asteriskLastReloadHours := 0
	asteriskLastReloadMinutes := 0
	asteriskLastReloadSeconds := 0

	re := regexp.MustCompile(``)

	for _, uptimeItem := range uptimeResults {
		if strings.Contains(uptimeItem, "System uptime") {

			re = regexp.MustCompile(`([0-9]+) year`)
			systemYears := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemYears) > 0 {
				systemUptimeYears, _ = strconv.Atoi(systemYears[0][1])
			}

			re = regexp.MustCompile(`([0-9]+) week`)
			systemWeeks := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemWeeks) > 0 {
				systemUptimeWeeks, _ = strconv.Atoi(systemWeeks[0][1])
			}

			re = regexp.MustCompile(`([0-9]+) day`)
			systemDays := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemDays) > 0 {
				systemUptimeDays, _ = strconv.Atoi(systemDays[0][1])
			}

			re = regexp.MustCompile(`([0-9]+) hour`)
			systemHours := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemHours) > 0 {
				systemUptimeHours, _ = strconv.Atoi(systemHours[0][1])
			}

			re = regexp.MustCompile(`([0-9]+) minute`)
			systemMinutes := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemMinutes) > 0 {
				systemUptimeMinutes, _ = strconv.Atoi(systemMinutes[0][1])
			}

			re = regexp.MustCompile(`([0-9]+) second`)
			systemSeconds := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemSeconds) > 0 {
				systemUptimeSeconds, _ = strconv.Atoi(systemSeconds[0][1])
			}
		}
		if strings.Contains(uptimeItem, "Last reload") {

			re = regexp.MustCompile(`([0-9]+) year`)
			reloadYears := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadYears) > 0 {
				asteriskLastReloadYears, _ = strconv.Atoi(reloadYears[0][1])
			}

			re = regexp.MustCompile(`([0-9]+) week`)
			reloadWeeks := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadWeeks) > 0 {
				asteriskLastReloadWeeks, _ = strconv.Atoi(reloadWeeks[0][1])
			}

			re = regexp.MustCompile(`([0-9]+) day`)
			reloadDays := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadDays) > 0 {
				asteriskLastReloadDays, _ = strconv.Atoi(reloadDays[0][1])
			}

			re = regexp.MustCompile(`([0-9]+) hour`)
			reloadHours := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadHours) > 0 {
				asteriskLastReloadHours, _ = strconv.Atoi(reloadHours[0][1])
			}

			re = regexp.MustCompile(`([0-9]+) minute`)
			reloadMinutes := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadMinutes) > 0 {
				asteriskLastReloadMinutes, _ = strconv.Atoi(reloadMinutes[0][1])
			}

			re = regexp.MustCompile(`([0-9]+) second`)
			reloadSeconds := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadSeconds) > 0 {
				asteriskLastReloadSeconds, _ = strconv.Atoi(reloadSeconds[0][1])
			}
		}
	}

	systemUptime = (systemUptimeYears * 31104000) + (systemUptimeWeeks * 604800) + (systemUptimeDays * 86400) + (systemUptimeHours * 3600) + (systemUptimeMinutes * 60) + systemUptimeSeconds

	asteriskLastReload = (asteriskLastReloadYears * 31104000) + (asteriskLastReloadWeeks * 604800) + (asteriskLastReloadDays * 86400) + (asteriskLastReloadHours * 3600) + (asteriskLastReloadMinutes * 60) + asteriskLastReloadSeconds

	//MFCR2 Channels
	mfcr2Result := command("mfcr2 show channels", s)

	mfcr2Result = strings.Replace(mfcr2Result, "\r\n", "\n", -1)
	mfcr2Results := strings.Split(mfcr2Result, "\n")

	mfcr2TotalChannels := 0
	mfcr2InuseChannels := 0
	mfcr2AvailableChannels := 0
	mfcr2BlockedChannels := 0

	if !strings.Contains(mfcr2Result, "No such command") {

		for _, mfcr2Item := range mfcr2Results {
			if strings.Contains(mfcr2Item, "IDLE") {
				mfcr2TotalChannels = mfcr2TotalChannels + 1
				mfcr2InuseChannels = mfcr2InuseChannels + 1
			}
			if strings.Contains(mfcr2Item, "ANSWER") {
				mfcr2TotalChannels = mfcr2TotalChannels + 1
				mfcr2AvailableChannels = mfcr2AvailableChannels + 1
			}
			if strings.Contains(mfcr2Item, "BLOCK") {
				mfcr2TotalChannels = mfcr2TotalChannels + 1
				mfcr2BlockedChannels = mfcr2BlockedChannels + 1
			}
		}
	}

	//Fields association
	fields := make(map[string]interface{})
	fields["current_call_volume"] = currentCallValue
	fields["pri_open_channels"] = priOpenchannels
	fields["sip_peers"] = sipPeersCount
	fields["sip_monitored_online"] = onlineMonitoredPeers
	fields["sip_monitored_offline"] = offlineMonitoredPeers
	fields["sip_unmonitored_online"] = onlineUnmonitoredPeers
	fields["sip_unmonitored_offline"] = offlineUnmonitoredPeers
	fields["iax2_peers_total"] = iaxPeersTotalCount
	fields["iax2_peers_online"] = iaxPeersOnlineCount
	fields["iax2_peers_offline"] = iaxPeersOfflineCount
	fields["iax2_peers_unmonitored"] = iaxPeersUnmonitoredCount
	fields["dahdi_trunks_total"] = dahdiTotalTrunks
	fields["dahdi_trunks_online"] = dahdiOnlineTrunks
	fields["dahdi_trunks_offline"] = dahdiOfflineTrunks
	fields["g729_total"] = g729TotalCount
	fields["g729_encoders"] = g729EncodersCount
	fields["g729_decoders"] = g729DecodersCount
	fields["system_uptime"] = systemUptime
	fields["last_reload"] = asteriskLastReload
	fields["mfcr2_channels_total"] = mfcr2TotalChannels
	fields["mfcr2_channels_in_use"] = mfcr2InuseChannels
	fields["mfcr2_channels_available"] = mfcr2AvailableChannels
	fields["mfcr2_channels_blocked"] = mfcr2BlockedChannels

	tags := make(map[string]string)

	s.x += 1.0
	acc.AddFields("asterisk", fields, tags)

	return nil
}

func init() {
	inputs.Add("asterisk", func() telegraf.Input { return &Asterisk{x: 0.0} })
}
