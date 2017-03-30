package gorush

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	// NotiferErrHTTPClient http client for reporter.
	NotiferErrHTTPClient *http.Client

	// PushErrNotifURL the value of the URL to which the sending push error report
	PushErrNotifURL *url.URL
)

type (
	ReportErrDTO struct {
		Plat  int    `json:"platforma"`
		Token string `json:"token"`
		Err   string `json:"error"`
	}
)

// InitReporter initialize the report service.
func InitReporter() {
	if len(PushConf.Core.PushErrNotif) == 0 {
		LogAccess.Info("Notification about push errors is disabled. See general log.")
		return
	}
	var err error
	PushErrNotifURL, err = url.Parse(PushConf.Core.PushErrNotif)
	if err != nil {
		LogAccess.Warn("Invalid push errors notification uri", err)
		LogAccess.Info("Notification about push errors is disabled. See general log.")
		return
	}

	// https://blog.cloudflare.com/content/images/2016/06/Timeouts-002.png
	// TODO: Delays from settings
	NotiferErrHTTPClient = &http.Client{
		Timeout: 1 * time.Minute,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			// TLSHandshakeTimeout:   10 * time.Second,
			// ResponseHeaderTimeout: 20 * time.Second,
			// ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			// TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
		},
	}
}

// ReportPushError send bug report information.
func ReportPushError(plat int, token, errstr string) {
	if PushErrNotifURL == nil {
		return
	}
	// TODO: Execute via queue
	// TODO: Add access key
	// TODO: Retry

	body := new(bytes.Buffer)
	reportURL := PushErrNotifURL.String()

	err := json.NewEncoder(body).
		Encode(ReportErrDTO{plat, token, errstr})
	if err != nil {
		LogError.Errorln("report: error decode", err)
		return
	}

	_, err = NotiferErrHTTPClient.Post(
		reportURL,
		"application/json",
		body,
	)
	if err != nil {
		LogError.Errorln("report: error sending", reportURL, err)
		return
	}
	// TODO: Resend if not 200
}
