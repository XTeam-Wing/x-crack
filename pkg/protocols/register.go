package protocols

import (
	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/XTeam-Wing/x-crack/pkg/protocols/telnet"
)

// RegisterAllProtocols 注册所有协议处理器
func RegisterAllProtocols() {
	brute.RegisterProtocolHandler("socks5", SOCKS5Brute)
	// 注册HTTP代理爆破处理器
	brute.RegisterProtocolHandler("http_proxy", HTTPProxyBrute)
	brute.RegisterProtocolHandler("ssh", SSHBrute)
	brute.RegisterProtocolHandler("ftp", FTPBrute)
	brute.RegisterProtocolHandler("telnet", telnet.TelnetBrute)
	brute.RegisterProtocolHandler("mysql", MySQLBrute)
	brute.RegisterProtocolHandler("postgresql", PostgreSQLBrute)
	brute.RegisterProtocolHandler("redis", RedisBrute)
	brute.RegisterProtocolHandler("mongodb", MongoDBBrute)
	brute.RegisterProtocolHandler("http", HTTPBrute)
	brute.RegisterProtocolHandler("https", HTTPSBrute)
	brute.RegisterProtocolHandler("smb", SMBBrute)
	brute.RegisterProtocolHandler("rdp", RDPBrute)
	brute.RegisterProtocolHandler("vnc", VNCBrute)
	brute.RegisterProtocolHandler("snmp", SNMPBrute)
	brute.RegisterProtocolHandler("imap", IMAPBrute)
	brute.RegisterProtocolHandler("pop3", POP3Brute)
	brute.RegisterProtocolHandler("smtp", SMTPBrute)
	brute.RegisterProtocolHandler("amqp", AMQPBrute)

}
