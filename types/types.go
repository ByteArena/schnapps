package types

type NICIface struct {
	Model string
}

type NICSocket struct {
	Connect string
}

type NICTap struct {
	Ifname string
}

type NICUser struct {
	DHCPStart string
	Net       string
}

type QMPServer struct {
	Protocol string
	Addr     string
}

type NICBridge struct {
	Bridge string
	MAC    string
}

type VMConfig struct {
	NICs          []interface{}
	Id            int
	ImageLocation string
	QMPServer     *QMPServer
	MegMemory     int
	CPUAmount     int
	CPUCoreAmount int
}

type VMMetadata map[string]string
