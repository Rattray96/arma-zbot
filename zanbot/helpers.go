package zanbot

import (
	"log"
	"strconv"
	"strings"
	"time"
)

type banData struct {
	name   string
	guid   string
	reason string
	file   string
}

type rconinfo struct {
	Host string
	Port string
	Pw   string
}

type jsonsettings struct {
	RemotExecFile string
	PublicVarFile string
	ScriptsFile   string
	Rcons         []rconinfo
	TestMode      bool
}

// get remote execution time of this line in the file.
// this converts the date time string in the line segment into a time.Time and returns.
func getReTime(s string) time.Time {
	// format testing.
	//fmt.Println("day:",s[0:2], "Month:",s[3:5],"year:",s[6:10],"h:",s[11:13],"m:",s[14:16],"s:",s[17:19])
	year, err := strconv.Atoi(s[6:10])
	if err != nil {
		log.Fatal(err)
	}
	month, err := strconv.Atoi(s[3:5])
	if err != nil {
		log.Fatal(err)
	}
	day, err := strconv.Atoi(s[0:2])
	if err != nil {
		log.Fatal(err)
	}
	hour, err := strconv.Atoi(s[11:13])
	if err != nil {
		log.Fatal(err)
	}

	min, err := strconv.Atoi(s[14:16])
	if err != nil {
		log.Fatal(err)
	}

	sec, err := strconv.Atoi(s[17:19])
	if err != nil {
		log.Fatal(err)
	}
	loc, _ := time.LoadLocation("Local")
	return time.Date(year, time.Month(month), day, hour, min, sec, 0, loc)

}

func getPlayerName(s string) string {
	pos := strings.Index(s, " - #0 \"")
	if pos == -1 {
		pos = strings.Index(s, " - Compile Block")
	}
	s = s[21:pos]
	pos = strings.LastIndex(s, "(")
	return string(s[0 : pos-1])
}

func getPlayerGUID(s string) string {
	extra := 32
	pos := strings.Index(s, " - #0 \"")
	if pos == -1 {
		pos = strings.Index(s, " - Compile Block")
	}
	if pos == -1 {
		panic("invalid pos in getplayerduid")
	}

	return s[pos-extra : pos]
}

func getBanReasonShort(s string) string {
	extra := 7
	pos := strings.Index(s, " - #0 \"")
	if pos == -1 {
		extra = 3
		pos = strings.Index(s, " - Compile Block")
	}

	if len(s) < pos+107 {
		return s[pos+extra : len(s)-1]
	}
	return s[pos+extra : pos+107]
}

func getBanReason(s string) string {
	extra := 7
	pos := strings.Index(s, " - #0 \"")
	if pos == -1 {
		extra = 3
		pos = strings.Index(s, " - Compile Block")
	}
	return s[pos+extra:]
}
