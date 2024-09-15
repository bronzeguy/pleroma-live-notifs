package main

import (
	"fmt"
	"net/http"
	"time"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"slices"
	"strings"
	"unicode/utf8"
)

var (
	TOKEN string
	INSTANCE string
)

const (
    htmlTagStart = 60 // Unicode `<`
    htmlTagEnd   = 62 // Unicode `>`
)

type notif struct {
	Account struct {
		Acct string `json:"acct"`
	} `json:"account"`
	Id string `json:"id"`
	Type string `json:"type"`
	Status struct {
		Content string `json:"content"`
		Visibility string `json:"visibility"`
		InReplyTo any `json:"in_reply_to_id"`
	} `json:"status"`
}

func request_notifs(c *http.Client, since_id string) []notif {
	var (
		err error
		req *http.Request
		notifs []notif
		since_id_int int
	)
	since_id_int, err = strconv.Atoi(since_id)
	if err != nil {
		panic(err)
	}
	if since_id_int == 0 {
		req, err = http.NewRequest("GET", fmt.Sprintf("https://%s/api/v1/notifications", INSTANCE), nil)
	} else {
		req, err = http.NewRequest("GET", fmt.Sprintf("https://%s/api/v1/notifications?since_id=%d", INSTANCE, since_id_int), nil)
	}
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer " + TOKEN)
	resp, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	/* using panic here made me use sprintf for the error i wanted to throw and it didn't look very good, but some consistency for error handling would probably be warranted anyway */
	if resp.StatusCode != 200 {
		fmt.Printf("bad status code: %d\n\n%s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}
	err = json.Unmarshal(body, &notifs)
	if err != nil {
		panic(err)
	}
	return notifs
}

/* Aggressively strips HTML tags from a string.
 * It will only keep anything between `>` and `<`.
 * I stole this from a nice man on Stack Overflow!!
 */
func stripHtmlTags(s string) string {
    // Setup a string builder and allocate enough memory for the new string.
    var builder strings.Builder
    builder.Grow(len(s) + utf8.UTFMax)

    in := false // True if we are inside an HTML tag.
    start := 0  // The index of the previous start tag character `<`
    end := 0    // The index of the previous end tag character `>`

    for i, c := range s {
        // If this is the last character and we are not in an HTML tag, save it.
        if (i+1) == len(s) && end >= start {
            builder.WriteString(s[end:])
        }

        // Keep going if the character is not `<` or `>`
        if c != htmlTagStart && c != htmlTagEnd {
            continue
        }

        if c == htmlTagStart {
            // Only update the start if we are not in a tag.
            // This make sure we strip out `<<br>` not just `<br>`
            if !in {
                start = i

                // Write the valid string between the close and start of the two tags.
                builder.WriteString(s[end:start])
            }
            in = true
            continue
        }
        // else c == htmlTagEnd
        in = false
        end = i + 1
    }
    s = builder.String()
    return s
}

func main() {
	since_id := "0"
	c := &http.Client{
		Timeout: 30 * time.Second,
	}
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s [details]\n\ndetails is a file that contains the instance your account is on and the authentication token for your account on that server, seperated by newlines\n", os.Args[0])
		os.Exit(1)
	}
	s, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	s2 := strings.Split(string(s), "\n")
	INSTANCE = s2[0]
	TOKEN = s2[1]
	for {
		time.Sleep(1 * time.Second)
		notifs := request_notifs(c, since_id)
		if len(notifs) == 0 {
			continue
		}
		slices.Reverse(notifs)
		for _,v := range(notifs) {
			switch v.Type {
			case "mention":
				if v.Status.InReplyTo == nil {
					fmt.Printf("%s tagged you in a %s context: %s\n\n", v.Account.Acct, v.Status.Visibility, stripHtmlTags(v.Status.Content))
				} else {
					fmt.Printf("%s replied to a post in a %s context: %s\n\n", v.Account.Acct, v.Status.Visibility, stripHtmlTags(v.Status.Content))
				}
			case "favourite":
				fmt.Printf("%s favorited your post: %s\n\n", v.Account.Acct, stripHtmlTags(v.Status.Content))
			case "follow":
				fmt.Printf("%s followed you\n\n", v.Account.Acct)
			case "reblog":
				fmt.Printf("%s retweeted your post: %s\n\n", v.Account.Acct, stripHtmlTags(v.Status.Content))
			default:
				fmt.Printf("unknown notification type %s\n\n", v.Type) /* /shrug */
			}
		}
		since_id = notifs[len(notifs)-1].Id
	}
}
