package utils

import (
	"os"
	"strconv"
	"strings"
)

var admins, helpers []uint64

func ParseBotAdmins() {
	for _, id := range strings.Split(os.Getenv("WORKER_BOT_ADMINS"), ",") {
		if parsed, err := strconv.ParseUint(id, 10, 64); err == nil {
			admins = append(admins, parsed)
		}
	}
}

func ParseBotHelpers() {
	for _, id := range strings.Split(os.Getenv("WORKER_BOT_HELPERS"), ",") {
		if parsed, err := strconv.ParseUint(id, 10, 64); err == nil {
			helpers = append(helpers, parsed)
		}
	}
}

func IsBotAdmin(id uint64) bool {
	for _, admin := range admins {
		if admin == id {
			return true
		}
	}

	return false
}

func IsBotHelper(id uint64) bool {
	if IsBotAdmin(id) {
		return true
	}

	for _, helper := range helpers {
		if helper == id {
			return true
		}
	}

	return false
}
