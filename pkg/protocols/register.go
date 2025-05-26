package protocols

import "github.com/XTeam-Wing/x-crack/pkg/brute"

// RegisterAllProtocols 注册所有协议处理器
func RegisterAllProtocols() {
	brute.RegisterProtocolHandler("ssh", SSHBrute)
	brute.RegisterProtocolHandler("ftp", FTPBrute)
	brute.RegisterProtocolHandler("telnet", TelnetBrute)
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
}
