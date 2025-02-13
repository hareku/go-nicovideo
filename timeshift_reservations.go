package nicovideo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-json-experiment/json"
	"golang.org/x/net/html"
)

type TimeshiftReservation struct {
	ProgramID string `json:"programId"`

	Program struct {
		Title    string `json:"title"`
		Schedule struct {
			BeginTime time.Time `json:"beginTime,format:'2006-01-02T15:04:05-0700'"`
			EndTime   time.Time `json:"endTime,format:'2006-01-02T15:04:05-0700'"`
			OpenTime  time.Time `json:"openTime,format:'2006-01-02T15:04:05-0700'"`
			Status    string    `json:"status"`
		} `json:"schedule"`
	} `json:"program"`

	SocialGroup struct {
		Name string `json:"name"`
	} `json:"socialGroup"`
}

// URL returns the URL of the timeshift reservation.
// e.g. https://live.nicovideo.jp/watch/lv123456789
func (r *TimeshiftReservation) URL() string {
	return fmt.Sprintf("https://live.nicovideo.jp/watch/%s", r.ProgramID)
}

func (c *Client) TimeshiftReservations(ctx context.Context) ([]*TimeshiftReservation, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://live.nicovideo.jp/embed/timeshift-reservations", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.AddCookie(&http.Cookie{
		Name:     "user_session",
		Value:    c.userSession,
		Domain:   ".nicovideo.jp",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	dataProps, err := extractHTMLDataProps(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("extract data-props: %w", err)
	}

	var props struct {
		Reservations struct {
			Reservations []*TimeshiftReservation `json:"reservations"`
		} `json:"reservations"`
	}
	if err := json.Unmarshal([]byte(dataProps), &props); err != nil {
		return nil, fmt.Errorf("unmarshal data-props: %w", err)
	}

	return props.Reservations.Reservations, nil
}

func extractHTMLDataProps(r io.Reader) (string, error) {
	tokenizer := html.NewTokenizer(r)
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return "", fmt.Errorf("data-props not found")
		case html.StartTagToken:
			token := tokenizer.Token()
			if token.Data == "script" {
				var hasID bool
				for _, attr := range token.Attr {
					if attr.Key == "id" && attr.Val == "embedded-data" {
						hasID = true
					}
				}
				if hasID {
					for _, attr := range token.Attr {
						if attr.Key == "data-props" {
							return html.UnescapeString(attr.Val), nil
						}
					}
				}
			}
		}
	}
}
