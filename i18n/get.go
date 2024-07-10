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

func LoadMessages() {
	for idx, locale := range Locales {
		path := fmt.Sprintf("./locale/%s.json", locale.IsoLongCode)

		data, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Failed to read locale %s: %s\n", locale.IsoShortCode, err.Error())

			if locale.IsoLongCode == "en-GB" { // Required
				panic(err)
			}
		}

		messages, err := parseCrowdInFile(data)
		if err != nil {
			fmt.Printf("Failed to parse locale: %s\n", err.Error())
			Locales[idx].Messages = make(map[MessageId]string)
			continue
		}

		Locales[idx].Messages = messages
	}
}

func SeedCoverage() {
	total := len(LocaleEnglish.Messages)

	for _, locale := range Locales {
		locale.Coverage = len(locale.Messages) * 100 / total
	}
}

func GetMessage(locale *Locale, id MessageId, format ...interface{}) string {
	if locale == nil {
		locale = LocaleEnglish
	}

	if locale.Messages == nil {
		if locale == LocaleEnglish {
			return fmt.Sprintf("Error: translations for language `%s` is missing", locale.IsoShortCode)
		}

		locale = LocaleEnglish // default to english
		return GetMessage(locale, id, format...)
	}

	value, ok := locale.Messages[id]
	if !ok || value == "" {
		if locale == LocaleEnglish {
			return fmt.Sprintf("error: translation for `%s` is missing", id)
		}

		return GetMessage(LocaleEnglish, id, format...) // default to English
	}

	return fmt.Sprintf(strings.Replace(value, "\\n", "\n", -1), format...)
}

func GetMessageFromGuild(guildId uint64, id MessageId, format ...interface{}) string {
	// TODO: Propagate context
	activeLanguage, err := dbclient.Client.ActiveLanguage.Get(context.Background(), guildId)
	if err != nil {
		sentry.Error(err)
	}

	if activeLanguage != "" {
		return GetMessage(MappedByIsoShortCode[activeLanguage], id, format...)
	}

	// check preferred locale
	preferredLocale, err := getPreferredLocale(guildId)
	if err != nil {
		if err != pgx.ErrNoRows {
			sentry.Error(err)
		}

		return GetMessage(LocaleEnglish, id, format...)
	}

	if preferredLocale == nil {
		return GetMessage(LocaleEnglish, id, format...)
	} else {
		language, ok := DiscordLocales[*preferredLocale]
		if !ok {
			language = LocaleEnglish
		}

		return GetMessage(language, id, format...)
	}
}

func getPreferredLocale(guildId uint64) (locale *string, err error) {
	query := `SELECT "data"->'preferred_locale' FROM guilds WHERE "guild_id" = $1;`
	err = cache.Client.QueryRow(context.Background(), query, guildId).Scan(&locale)
	return
}

func parseCrowdInFile(data []byte) (map[MessageId]string, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}

	return parseCrowdInData("", parsed), nil
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
			if s == "" {
				continue
			}

			parsed[MessageId(newPath)] = s
		} else if m, ok := value.(map[string]interface{}); ok {
			// TODO: Pass the map down directly
			for k, v := range parseCrowdInData(newPath, m) {
				if v == "" {
					continue
				}

				parsed[k] = v
			}
		} else {
			panic(fmt.Sprintf("key %s.%s has unknown type", path, key))
		}
	}

	return parsed
}
