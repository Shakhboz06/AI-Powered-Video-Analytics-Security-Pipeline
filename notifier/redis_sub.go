package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	"video-analytics-pipe/config"

	"github.com/redis/go-redis/v9"
)

type Alert struct {
	Camera     string     `json:"camera"`
	ZoneID     *int64     `json:"zone_id"`
	ZoneName   string     `json:"zone_name"`
	TrackerID  int64      `json:"tracker_id"`
	BoundBox   [4]float64 `json:"bound_box"`
	Label      string     `json:"label"`
	RecordedAt time.Time  `json:"recorded_at"`
	AlertType  string     `json:"alert_type"`
	Severity   string     `json:"severity"`
}

func SubToRedis(rdb *redis.Client) error {

	mailHost := config.GetString("SMTP_HOST", "")
	mailPort := config.GetString("SMTP_PORT", "")
	mailUsername := config.GetString("SMTP_USERNAME", "")
	mailPassword := config.GetString("SMTP_PASSWORD", "")
	mailFrom := config.GetString("SMTP_FROM", "")
	mailRecipients := strings.Split(config.GetString("NOTIFIER_RECIPIENTS", ""), ",")

	mailer := NewMailer(mailHost, mailPort, mailUsername, mailPassword, mailFrom, mailRecipients)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := rdb.Subscribe(ctx, "alerts")
	defer sub.Close()

	ch := sub.Channel()
	for msg := range ch {

		var alert Alert
		if err := json.Unmarshal([]byte(msg.Payload), &alert); err != nil {
			log.Printf("unmarshal alert: %v", err)
			continue
		}

		if alert.Severity == "warning" {
			err := mailer.SendMail(fmt.Sprintf("%v alert in %s", alert.AlertType, alert.ZoneName), fmt.Sprintf("camera: %s, zoneID: %v, trackerID: %v, object: %v, time: %v", alert.Camera, *alert.ZoneID, alert.TrackerID, alert.Label, alert.RecordedAt))
			if err != nil {
				log.Printf("sending failed: %v", err)
				continue
			}
		}

		if alert.Severity == "info" {
			err := mailer.SendMail(fmt.Sprintf("%v alert in %s", alert.AlertType, alert.ZoneName), fmt.Sprintf("camera: %s, zoneID: %v, trackerID: %v, object: %v, time: %v", alert.Camera, *alert.ZoneID, alert.TrackerID, alert.Label, alert.RecordedAt))
			if err != nil {
				log.Printf("sending failed: %v", err)
				continue
			}
		}

		if alert.Severity == "critical" {
			err := mailer.SendMail(fmt.Sprintf("%v alert in %s", alert.AlertType, alert.Camera), fmt.Sprintf("camera: %s, trackerID: %v, object: %v, time: %v", alert.Camera, alert.TrackerID, alert.Label, alert.RecordedAt))
			if err != nil {
				log.Printf("sending failed: %v", err)
				continue
			}
		}

	}
	return nil
}
