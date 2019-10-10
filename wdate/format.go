package wdate

import (
	"strconv"
	"strings"
	"time"
)

// Format return current or given date with php format as as string
func Format(dateStr string, ts ...time.Time) string {

	var timestamp time.Time

	if len(ts) > 0 {
		timestamp = ts[0]
	} else {
		timestamp = time.Now()
	}

	// nan mais sérieusement https://yourbasic.org/golang/dateStr-parse-string-time-date-example/

	// Macros spécifiques
	minutes, _ := strconv.Atoi(timestamp.Format("4"))
	tensMinutes := minutes / 10
	dateStr = strings.Replace(dateStr, "{ti}", strconv.Itoa(tensMinutes), -1) // Dizaines de minutes

	dateStr = strings.Replace(dateStr, "Y", timestamp.Format("2006"), -1) // Année sur 4 chiffres
	dateStr = strings.Replace(dateStr, "y", timestamp.Format("06"), -1)   // Année sur 2 chiffres

	dateStr = strings.Replace(dateStr, "F", timestamp.Format("January"), -1) // Mois, textuel, version longue; en anglais, comme January ou December
	dateStr = strings.Replace(dateStr, "m", timestamp.Format("01"), -1)      // Mois au dateStr numérique, avec zéros initiaux
	dateStr = strings.Replace(dateStr, "M", timestamp.Format("Jan"), -1)     // Mois, en trois lettres, en anglais
	dateStr = strings.Replace(dateStr, "n", timestamp.Format("1"), -1)       // Mois sans les zéros initiaux

	dateStr = strings.Replace(dateStr, "d", timestamp.Format("02"), -1)  // Jour du mois, sur deux chiffres (avec un zéro initial)
	dateStr = strings.Replace(dateStr, "D", timestamp.Format("Mon"), -1) // Jour de la semaine, en trois lettres
	dateStr = strings.Replace(dateStr, "j", timestamp.Format("2"), -1)   // Jour du mois sans les zéros initiaux

	dateStr = strings.Replace(dateStr, "H", timestamp.Format("15"), -1) // Heure, au format 24h, avec les zéros initiaux
	dateStr = strings.Replace(dateStr, "g", timestamp.Format("3"), -1)  // Heure, au format 12h, sans les zéros initiaux
	dateStr = strings.Replace(dateStr, "h", timestamp.Format("03"), -1) // Heure, au format 12h, avec les zéros initiaux

	dateStr = strings.Replace(dateStr, "i", timestamp.Format("04"), -1) // Minutes avec les zéros initiaux

	dateStr = strings.Replace(dateStr, "s", timestamp.Format("05"), -1) // Secondes, avec zéros initiaux

	return dateStr
}
