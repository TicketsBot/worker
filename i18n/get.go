package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/cache"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/jackc/pgx/v4"
	"io/ioutil"
	"strings"
)

var messages map[Language]map[MessageId]string
var coverage map[Language]int

func LoadMessages() {
	messages = make(map[Language]map[MessageId]string)

	for locale, language := range FullLocales {
		path := fmt.Sprintf("./locale/%s.json", locale)

		data, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Failed to read locale %s: %s\n", locale, err.Error())

			if locale == "en-GB" { // Required
				panic(err)
			}
		}

		messages[language] = parseCrowdInFile(data)
	}
}

func SeedCoverage() {
	coverage = make(map[Language]int)

	total := len(messages[English])

	for _, language := range FullLocales {
		translated := len(messages[language])
		coverage[language] = translated * 100 / total
	}
}

func GetCoverage(language Language) int {
	coverage, ok := coverage[language]
	if ok {
		return coverage
	} else {
		return 0
	}
}

func GetMessage(language Language, id MessageId, format ...interface{}) string {
	if messages[language] == nil {
		if language == English {
			return fmt.Sprintf("Error: translations for language `%s` is missing", language)
		}

		language = English // default to english
		return GetMessage(language, id, format...)
	}

	value, ok := messages[language][id]
	if !ok || value == "" {
		if language == English {
			return fmt.Sprintf("error: translation for `%s` is missing", id)
		}

		return GetMessage(English, id, format...) // default to English
	}

	return fmt.Sprintf(strings.Replace(value, "\\n", "\n", -1), format...)
}

func GetMessageFromGuild(guildId uint64, id MessageId, format ...interface{}) string {
	activeLanguage, err := dbclient.Client.ActiveLanguage.Get(guildId)
	if err != nil {
		sentry.Error(err)
	}

	if activeLanguage != "" {
		return GetMessage(Language(activeLanguage), id, format...)
	}

	// check preferred locale
	preferredLocale, err := getPreferredLocale(guildId)
	if err != nil {
		if err != pgx.ErrNoRows {
			sentry.Error(err)
		}

		return GetMessage(English, id, format...)
	}

	if preferredLocale == nil {
		return GetMessage(English, id, format...)
	} else {
		language, ok := DiscordLocales[*preferredLocale]
		if !ok {
			language = English
		}

		return GetMessage(language, id, format...)
	}
}

func getPreferredLocale(guildId uint64) (locale *string, err error) {
	query := `SELECT "data"->'preferred_locale' FROM guilds WHERE "guild_id" = $1;`
	err = cache.Client.QueryRow(context.Background(), query, guildId).Scan(&locale)
	return
}

func parseCrowdInFile(data []byte) map[MessageId]string {
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		fmt.Printf("Failed to parse locale: %s\n", err.Error())
		return nil
	}

	return parseCrowdInData("", parsed)
}

func parseCrowdInData(path string, data map[string]interface{}) map[MessageId]string {
	parsed := make(map[MessageId]string)

	for key, value := range data {
		var newPath string
		if key == "" {
			newPath = path
		} else if path == "" {
			newPath = key
		} else {
			newPath = fmt.Sprintf("%s.%s", path, key)
		}

		s, ok := value.(string)
		if ok {
			parsed[MessageId(newPath)] = s
		} else if m, ok := value.(map[string]interface{}); ok {
			// TODO: Pass the map down directly
			for k, v := range parseCrowdInData(newPath, m) {
				parsed[k] = v
			}
		} else {
			panic(fmt.Sprintf("key %s.%s has unknown type", path, key))
		}
	}

	return parsed
}
