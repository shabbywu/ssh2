package db

const (
	// AuthMethod Index
	AuthMethodIndexPattern  = "kind:AuthMethod:id:*"
	AuthMethodIdIndexName   = "AuthMethod:id"
	AuthMethodNameIndexName = "AuthMethod:name"
	// ClientConfig Index
	ClientConfigIndexPattern  = "kind:ClientConfig:id:*"
	ClientConfigIdIndexName   = "ClientConfig:id"
	ClientConfigNameIndexName = "ClientConfig:name"
	// ServerConfig Index
	ServerConfigIndexPattern  = "kind:ServerConfig:id:*"
	ServerConfigIdIndexName   = "ServerConfig:id"
	ServerConfigNameIndexName = "ServerConfig:name"
	// Session Index
	SessionIndexPattern  = "kind:Session:id:*"
	SessionIdIndexName   = "Session:id"
	SessionNameIndexName = "Session:name"
	SessionTagIndexName  = "Session:tag"

	// MetaData Field
	MetaDataKey = "MetaData"
)

var (
	// 数据库路径
	DbPath = "~/.ssh/ssh2/db.bin"
)
