// Statup
// Copyright (C) 2018.  Hunter Long and the project contributors
// Written by Hunter Long <info@socialeck.com> and the project contributors
//
// https://github.com/hunterlong/statup
//
// The licenses for most software and other practical works are designed
// to take away your freedom to share and change the works.  By contrast,
// the GNU General Public License is intended to guarantee your freedom to
// share and change all versions of a program--to make sure it remains free
// software for all its users.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package notifiers

import (
	"bytes"
	"fmt"
	"github.com/hunterlong/statup/core/notifier"
	"github.com/hunterlong/statup/types"
	"github.com/hunterlong/statup/utils"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	WEBHOOK_METHOD = "webhook"
)

type Webhook struct {
	*notifier.Notification
}

var webhook = &Webhook{&notifier.Notification{
	Method:      WEBHOOK_METHOD,
	Title:       "HTTP Webhook",
	Description: "Send a custom HTTP request to a specific URL with your own body, headers, and parameters",
	Author:      "Hunter Long",
	AuthorUrl:   "https://github.com/hunterlong",
	Delay:       time.Duration(1 * time.Second),
	Form: []notifier.NotificationForm{{
		Type:        "text",
		Title:       "HTTP Endpoint",
		Placeholder: "http://webhookurl.com/JW2MCP4SKQP",
		SmallText:   "Insert the URL for your HTTP Requests",
		DbField:     "Host",
		Required:    true,
	}, {
		Type:        "text",
		Title:       "HTTP Method",
		Placeholder: "POST",
		SmallText:   "Choose a HTTP method for example: GET, POST, DELETE, or PATCH",
		DbField:     "Var1",
		Required:    true,
	}, {
		Type:        "textarea",
		Title:       "HTTP Body",
		Placeholder: `{"service_id": "%s.Id", "service_name": "%s.Name"}`,
		SmallText:   "Optional HTTP body for a POST request. You can insert variables into your body request.<br>%service.Id, %service.Name<br>%failure.Issue",
		DbField:     "Var2",
	}, {
		Type:        "text",
		Title:       "Content Type",
		Placeholder: `application/json`,
		SmallText:   "Optional content type for example: application/json or text/plain",
		DbField:     "api_key",
	}, {
		Type:        "text",
		Title:       "Header",
		Placeholder: "Authorization=Token12345",
		SmallText:   "Optional Headers for request use format: KEY=Value,Key=Value",
		DbField:     "api_secret",
	},
	}}}

// DEFINE YOUR NOTIFICATION HERE.
func init() {
	err := notifier.AddNotifier(webhook)
	if err != nil {
		panic(err)
	}
}

// Send will send a HTTP Post to the Webhook API. It accepts type: string
func (w *Webhook) Send(msg interface{}) error {
	message := msg.(string)
	_, err := w.run(message)
	return err
}

func (w *Webhook) Select() *notifier.Notification {
	return w.Notification
}

func replaceBodyText(body string, s *types.Service, f *types.Failure) string {
	if s != nil {
		body = strings.Replace(body, "%service.Name", s.Name, -1)
		body = strings.Replace(body, "%service.Id", utils.ToString(s.Id), -1)
	}
	if f != nil {
		body = strings.Replace(body, "%failure.Issue", f.Issue, -1)
	}
	return body
}

func (w *Webhook) run(body string) (*http.Response, error) {
	utils.Log(1, fmt.Sprintf("sending body: '%v' to %v as a %v request", body, w.Host, w.Var1))
	client := new(http.Client)
	client.Timeout = time.Duration(10 * time.Second)
	var buf *bytes.Buffer
	buf = bytes.NewBuffer(nil)
	if w.Var2 != "" {
		buf = bytes.NewBuffer([]byte(w.Var2))
	}
	req, err := http.NewRequest(w.Var1, w.Host, buf)
	if err != nil {
		return nil, err
	}
	if w.ApiSecret != "" {
		splitArray := strings.Split(w.ApiSecret, ",")
		for _, a := range splitArray {
			split := strings.Split(a, "=")
			req.Header.Add(split[0], split[1])
		}
	}
	if w.ApiSecret != "" {
		req.Header.Add("Content-Type", w.ApiSecret)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (w *Webhook) OnTest() error {
	service := &types.Service{
		Id:             1,
		Name:           "Interpol - All The Rage Back Home",
		Domain:         "https://www.youtube.com/watch?v=-u6DvRyyKGU",
		ExpectedStatus: 200,
		Interval:       30,
		Type:           "http",
		Method:         "GET",
		Timeout:        20,
		LastStatusCode: 404,
		Expected:       "test example",
		LastResponse:   "<html>this is an example response</html>",
		CreatedAt:      time.Now().Add(-24 * time.Hour),
	}
	body := replaceBodyText(w.Var2, service, nil)
	resp, err := w.run(body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	utils.Log(1, fmt.Sprintf("webhook notifier received: '%v'", string(content)))
	return err
}

// OnFailure will trigger failing service
func (w *Webhook) OnFailure(s *types.Service, f *types.Failure) {
	msg := replaceBodyText(w.Var2, s, f)
	webhook.AddQueue(msg)
	w.Online = false
}

// OnSuccess will trigger successful service
func (w *Webhook) OnSuccess(s *types.Service) {
	if !w.Online {
		msg := replaceBodyText(w.Var2, s, nil)
		webhook.AddQueue(msg)
	}
	w.Online = true
}

// OnSave triggers when this notifier has been saved
func (w *Webhook) OnSave() error {
	return nil
}