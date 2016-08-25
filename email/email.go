package email

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Clever/http-science/config"
	"github.com/Clever/http-science/science"
	"github.com/keighl/mandrill"
)

// SendEmail sends email to the address in the payload with the results
func SendEmail(payload *config.Payload, duration time.Duration, res science.Results) error {
	message := &mandrill.Message{}
	message.AddRecipient(payload.Email, "", "to")

	science.Res.Mutex.Lock()
	message.GlobalMergeVars = mandrill.MapToVars(map[string]interface{}{
		"TYPE":       payload.JobType,
		"REQS":       strconv.Itoa(res.Reqs),
		"NUM_DIFFS":  strconv.Itoa(res.Diffs),
		"DIFFS_MAP":  fmt.Sprintf("%#v", res.Codes),
		"DIFFS_FILE": payload.DiffLoc,
		"TIME":       duration.String(),
	})
	science.Res.Mutex.Unlock()
	templateName := "http-science-results"

	mandrillClient := mandrill.ClientWithKey(os.Getenv("MANDRILL_KEY"))
	_, err := mandrillClient.MessagesSendTemplate(message, templateName, map[string]string{})
	return err
}
