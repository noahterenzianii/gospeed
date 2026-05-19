package app

import (
	"fmt"
	"time"

	"github.com/noahterenzianii/gospeed/internal/endpoints"
	"github.com/noahterenzianii/gospeed/internal/speedtest"
)

func Run() error {
	info, err := endpoints.FetchClientInfo()
	if err != nil {
		return err
	}
	fmt.Printf("Testing from: %s (%s) %s - %s (%s)\n", info.IP, info.Org, info.Country, info.City, info.Region)

	servers, err := endpoints.FetchServers("https://librespeed.org/backend-servers/servers.php")
	if err != nil {
		return err
	}
	best := endpoints.FindBestServer(servers)
	if best == nil {
		fmt.Println("No server reachable")
		return nil
	}
	// fmt.Printf("Best Server: %v\n", best.Latency)

	pingURL := best.URL(best.PingURL)
	latency, err := endpoints.MeasureLatency(pingURL, 200)
	if err != nil {
		return err
	}
	fmt.Printf("Ping: %v\n", latency.Ping)
	fmt.Printf("jitter: %v\n", latency.Jitter)

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
			fmt.Printf("\rUpload: %.2f Mbit/s", currentMbps)
		})
	if err != nil {
		return err
	}
	fmt.Printf("\nUpload speed: %.2f Mbit/s\n", uploadSpeed)
	return nil
}
