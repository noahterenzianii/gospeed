package app

import (
	"fmt"
	"time"

	"github.com/noahterenzianii/gospeed/internal/endpoints"
	"github.com/noahterenzianii/gospeed/internal/speedtest"
)

func Run() error {
	servers, err := endpoints.FetchServers("https://librespeed.org/backend-servers/servers.php")
	if err != nil {
		return err
	}
	best := endpoints.FindBestServer(servers)
	if best == nil {
		fmt.Println("No server reachable")
		return nil
	}

	fmt.Printf("Best Server: %v\n", best.Latency)

	pingURL := best.URL(best.PingURL)
	jitter, err := endpoints.MeasureJitter(pingURL, 200)
	if err != nil {
		return err
	}
	fmt.Printf("jitter: %v\n", jitter)

	downloadURL := best.URL(best.DlURL)
	downloadSpeed, err := speedtest.MeasureDownload(downloadURL, 10*time.Second, 4,
		func(currentMbps float64) {
			fmt.Printf("\rDownload: %.2f Mbit/s", currentMbps)
		})
	if err != nil {
		return err
	}
	fmt.Printf("\nDownload speed: %.2f Mbit/s\n", downloadSpeed)

	uploadURL := best.URL(best.UlURL)
	uploadSpeed, err := speedtest.MeasureUpload(uploadURL, 10*time.Second, 4,
		func(currentMbps float64) {
			fmt.Printf("\rDownload: %.2f Mbit/s", currentMbps)
		})
	if err != nil {
		return err
	}
	fmt.Printf("\nUpload speed: %.2f Mbit/s\n", uploadSpeed)
	return nil
}
