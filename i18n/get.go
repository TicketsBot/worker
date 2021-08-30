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

		var parsed map[MessageId]string
		if err := json.Unmarshal(data, &parsed); err != nil {
			fmt.Printf("Failed to parse locale %s: %s\n", locale, err.Error())

			if locale == "en-GB" { // Required
				panic(err)
			}
		}

		messages[language] = parsed
	}
}

func GetMessage(language Language, id MessageId, format ...interface{}) string {
	if messages[language] == nil {
		if language == English {
			return fmt.Sprintf("Error: translation for %s is missing", id)
		}

		language = English // default to english
		return GetMessage(language, id, format...)
	}

	value, ok := messages[language][id]
	if !ok || value == "" {
		if language == English {
			return fmt.Sprintf("error: translation for %d is missing", id)
		}

		language = English // default to english
		return GetMessage(language, id, format...)
	}

	return fmt.Sprintf(strings.Replace(value, "\\n", "\n", -1), format...)
}

func GetMessageFromGuild(guildId uint64, id MessageId, format ...interface{}) string {
	language := English
	_language, err := dbclient.Client.ActiveLanguage.Get(guildId)
	if err == nil {
		language = Language(_language)
	} else {
		sentry.Error(err)
	}

	if language == "" {
		// check preferred locale
		preferredLocale, err := getPreferredLocale(guildId)
		if err == nil {
			if preferredLocale == nil {
				language = English
			} else {
				var ok bool
				language, ok = DiscordLocales[*preferredLocale]

				if !ok {
					language = English
				}
			}
		} else {
			language = English

			if err != pgx.ErrNoRows {
				sentry.Error(err)
			}
		}
	}

	return GetMessage(language, id, format...)
}

func getPreferredLocale(guildId uint64) (locale *string, err error) {
	query := `SELECT "data"->'preferred_locale' FROM guilds WHERE "guild_id" = $1;`
	err = cache.Client.QueryRow(context.Background(), query, guildId).Scan(&locale)
	return
}
