package util

const (
	Cross = string(rune(0x274c))
	Tick  = string(rune(0x2714))
	Wait  = string(rune(0x267B))
	Run   = string(rune(0x1F3C3))
	Warn  = string(rune(0x26A0))
	Lock  = string(rune(0x1F512))
	Globe = string(rune(0x1F310))
)

func Printf(format string, a ...interface{}) {
	Output.Infof(format, a...)
}

func Fatalf(format string, a ...interface{}) {
	Output.Fatalf(format, a...)
}
