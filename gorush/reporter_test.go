package gorush

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"github.com/appleboy/gorush/config"
	"github.com/stretchr/testify/assert"
)

// TODO: more tests

func TestPushToAndroidFail_CheckReport(t *testing.T) {
	var done = make(chan *ReportErrDTO)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		report := &ReportErrDTO{}
		json.NewDecoder(r.Body).Decode(report)
		done <- report
	}))
	defer ts.Close()

	// config
	PushConf = config.BuildDefaultPushConf()
	PushConf.Core.PushErrNotif = ts.URL
	PushConf.Ios.Enabled = true
	PushConf.Ios.KeyPath = "../../backend/gorush/ios/apns-dev-cert.pem"

	// init
	InitLog()
	InitAPNSClient()
	InitAppStatus()
	InitReporter()

	// main
	req := PushNotification{
		Tokens:   []string{"f6c833591711ef13bec877a1773b3df791aafcc5268e1399cf7da30a0442a49f"},
		Platform: 1,
		Message:  "Welcome",
	}

	isError := PushToIOS(req)
	assert.True(t, isError)

	select {
	case report := <-done:
		assert.Equal(t, PlatFormIos, report.Plat, "expected platform from ios")
		assert.Equal(t, "f6c833591711ef13bec877a1773b3df791aafcc5268e1399cf7da30a0442a49f", report.Token, "expected another token")
		assert.Equal(t, "Unregistered", report.Err)
	case <-time.After(time.Millisecond * 1):
		assert.Fail(t, "report timeout")
	}

}
