package counters

// counters allows to get/set/increment or decrement some values
type Counters interface {
	Increment(name string)
	Decrement(name string)
	Get(name string) int64
	Set(name string, value int64)
}
