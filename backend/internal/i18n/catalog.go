package i18n

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type Key string

const (
	KeyTestNotificationTitle   Key = "notification.test.title"
	KeyTestNotificationMessage Key = "notification.test.message"
	KeyPrinterStarted          Key = "printer.started"
	KeyPrinterCompleted        Key = "printer.completed"
)

type Catalog struct {
	messages map[Key]string
}

//go:embed locales/pt-BR.json
var ptBRMessages []byte

func LoadPtBR() (Catalog, error) {
	messages := make(map[Key]string)

	if err := json.Unmarshal(ptBRMessages, &messages); err != nil {
		return Catalog{}, fmt.Errorf("decode pt-BR translations: %w", err)
	}

	for _, key := range []Key{
		KeyTestNotificationTitle,
		KeyTestNotificationMessage,
		KeyPrinterStarted,
		KeyPrinterCompleted,
	} {
		if messages[key] == "" {
			return Catalog{}, fmt.Errorf("missing pt-BR translation for %q", key)
		}
	}

	return Catalog{messages: messages}, nil
}

func (c Catalog) Text(key Key) string {
	return c.messages[key]
}
