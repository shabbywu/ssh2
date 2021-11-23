package db

type MetaDataID struct {
	AuthMethod   int
	ClientConfig int
	ServerConfig int
	Session      int
}

type MetaData struct {
	ID MetaDataID
}
