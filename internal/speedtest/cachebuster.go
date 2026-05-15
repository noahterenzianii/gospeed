package speedtest

import (
	"strconv"
	"strings"
	"time"
)

// make sure that every respone isnt cached
func withCacheBuster(url string, workerID, run int) string {
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	return url + sep + "n=" + strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + strconv.Itoa(workerID) + "_" + strconv.Itoa(run)
}
