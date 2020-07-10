package i18n

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/cache"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/jackc/pgx/v4"
	"strings"
)

var messages map[translations.Language]map[translations.MessageId]string

func LoadMessages(db *database.Database) (err error) {
	messages, err = db.Translations.GetAll()
	return
}

func GetMessage(language translations.Language, id translations.MessageId) string {
	if messages[language] == nil {
		if language == translations.English {
			return "error: lang en is missing"
		}

		language = translations.English // default to english
		return GetMessage(language, id)
	}

	value, ok := messages[language][id]
	if !ok || value == "" {
		if language == translations.English {
			return fmt.Sprintf("error: translation for %d is missing", id)
		}

		language = translations.English // default to english
		return GetMessage(language, id)
	}

	return strings.Replace(value, "\\n", "\n", -1)
}

func GetMessageFromGuild(guildId uint64, id translations.MessageId) string {
	language, err := dbclient.Client.ActiveLanguage.Get(guildId)
	if err != nil {
		sentry.Error(err)
		language = translations.English
	}

	if language == "" {
		// check preferred locale
		preferredLocale, err := getPreferredLocale(guildId)
		if err == nil {
			if preferredLocale == nil {
				language = translations.English
			} else {
				var ok bool
				language, ok = translations.Locales[*preferredLocale]

				if !ok {
					language = translations.English
				}
			}
		} else {
			language = translations.English

			if err != pgx.ErrNoRows {
				sentry.Error(err)
			}
		}
	}

	return GetMessage(language, id)
}

func getPreferredLocale(guildId uint64) (locale *string, err error) {
	query := `SELECT "data"->'preferred_locale' FROM guilds WHERE "guild_id" = $1;`
	err = cache.Client.QueryRow(context.Background(), query, guildId).Scan(&locale)
	return
}
