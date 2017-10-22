package qmp

var (
	START = 44400

	// There is technical limitation here. Only qmp.MAX VMs can run at the
	// same time
	MAX = 99

	inc = 0
)

func GetNextPort() int {
	port := START + (inc % (MAX + 1))
	inc++

	return port
}
