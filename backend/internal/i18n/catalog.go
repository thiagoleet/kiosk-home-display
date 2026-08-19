package i18n

type Catalog struct {
	TestNotificationTitle   string
	TestNotificationMessage string
	PrinterStarted          string
	PrinterCompleted        string
}

func PtBR() Catalog {
	return Catalog{
		TestNotificationTitle:   "Notificação de teste",
		TestNotificationMessage: "Esta notificação veio do backend Go.",
		PrinterStarted:          "Impressão iniciada",
		PrinterCompleted:        "Impressão concluída",
	}
}
