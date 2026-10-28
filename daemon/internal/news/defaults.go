package news

import (
	"database/sql"
	"fmt"
)

// defaultSources are the feeds a database starts with. They are written once,
// by the migration step SeedDefaultSources, and from then on the table is what
// counts: editing this list reaches a fresh database only. A default that must
// also reach boxes already installed needs a migration step of its own.
//
// Every feed URL here was checked to serve RSS. EPTV is left out until a feed
// of its own is found: its domain no longer resolves, and its stories are
// published through g1's regional feeds, which are g1's and not EPTV's. A
// website address is never listed in place of a feed.
var defaultSources = []Source{
	{
		Name:     "Agência Brasil",
		FeedURL:  "https://agenciabrasil.ebc.com.br/feed/ultimasnoticias/feed.xml",
		Category: DefaultCategory,
	},
	{
		Name:     "G1",
		FeedURL:  "https://g1.globo.com/rss/g1/",
		Category: DefaultCategory,
	},
	{
		Name:     "BBC News Brasil",
		FeedURL:  "https://feeds.bbci.co.uk/portuguese/rss.xml",
		Category: DefaultCategory,
	},
	{
		Name:     "ACidade ON Ribeirão Preto",
		FeedURL:  "https://www.acidadeon.com/ribeiraopreto/feed/",
		Category: DefaultCategory,
	},
}

// SeedDefaultSources adds the default sources, enabled, to news_sources. A
// feed URL already in the table is left exactly as it is — its name, category
// and enabled flag included — so a box upgraded with sources of its own keeps
// them, and no article loses the source it points to.
//
// It touches no network: whether a feed answers is the scheduler's concern.
func SeedDefaultSources(tx *sql.Tx) error {
	const insert = `
		INSERT INTO news_sources (name, feed_url, category, enabled)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(feed_url) DO NOTHING
	`

	for _, source := range defaultSources {
		if _, err := tx.Exec(
			insert,
			source.Name,
			source.FeedURL,
			source.Category,
		); err != nil {
			return fmt.Errorf(
				"failed to seed news source %q: %w",
				source.Name,
				err,
			)
		}
	}

	return nil
}
