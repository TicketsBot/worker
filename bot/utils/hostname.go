package utils

import "os"

const DefaultServiceName = "worker"

func GetServiceName() string {
	if _, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST"); !ok {
		return DefaultServiceName
	}

	hostname, err := os.Hostname()
	if err != nil || len(hostname) == 0 {
		return DefaultServiceName
	} else {
		var dashCount int

		for i := len(hostname) - 1; i >= 0; i-- {
			if hostname[i] == '-' {
				dashCount++
			}

			if dashCount == 2 {
				return hostname[:i]
			}
		}

		return DefaultServiceName
	}
}
