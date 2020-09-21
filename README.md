AstTelegraf
===================
#Asterisk Telegraf Integration (NON OFFICIAL)
<p align="center">
  <img width="400px" src="https://raw.githubusercontent.com/mafairnet/asterisk-telegraf-integration/master/resources/screenshot.png">
</p>

About
-----------
#### Objective: Provide a Open Souce Monitoring Alternative to Monitor Asterisk VoIP Open Source Software.

Telegraf integration (NON OFFICIAL) to Monitor Asterisk VoIP Open Source Software with basica metrics like calls, SIP peers, IAX2 peers, Channels, etc.

Prerequisites
-----------
- InfluxDB Server
- Grafana Server
- Asterisk Server

Usage
-----------
1. Get Telegraf Agent binary with Asterisk Input Plugin added.
```
wget https://raw.githubusercontent.com/mafairnet/asterisk-telegraf-integration/archive/master.zip
```
2. Uncompress File
```
tar -xvf asterisk-telegraf-integration-master.zip
```
3. Set write permissions to binary
```
cd asterisk-telegraf-integration-master
chmod +x telegraf
```
4. Edit telegraf Agent Configuration
```
# Collects basic metrics dfrom Asterisk Open Source VoIP Software
[agent]
  hostname = ""
[[outputs.influxdb]]
  urls = ["http://0.0.0.0:8086"]
[[inputs.asterisk]]
  ##Sample Config
  asterisk_ip = "0.0.0.0"
  ami_port = 5038
  ami_user = "user"
  ami_password = "password"
```
5. Run Telegraf Agent
```
../telegraf -config telegraf.conf
```
5. Go to your grafana and cerate a new dashboard and configure to your needs.