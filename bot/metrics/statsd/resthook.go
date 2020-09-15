package statsd

func RestHook(string) {
	Client.IncrementKey(REST)
}
