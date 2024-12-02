package fee

const minerFeeSatoshisPerByte = 25

func Get(rawTx string) int {
	return len([]byte(rawTx)) * satoshisPerByte
}
