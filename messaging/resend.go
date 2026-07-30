package messaging

import (
	"github.com/resend/resend-go/v3"
	"github.com/skb1129/go-utils/config"
	"github.com/skb1129/go-utils/logs"
)

type Resend struct {
	client *resend.Client
}

func NewResend() *Resend {
	logger := logs.GetLogger()
	apiKey := config.GetString("resend.apiKey")
	if apiKey == "" {
		logger.Fatal("Resend email configuration is incomplete")
	}
	client := resend.NewClient(apiKey)
	return &Resend{client}
}

func (c *Resend) SendEmails(templateID string, toEmails []string, dataList []map[string]any) map[string]int {
	if len(toEmails) == 0 || len(dataList) == 0 {
		return nil
	}
	failureCount := 0
	successCount := 0
	for i := 0; i < len(toEmails); i += 50 {
		end := min(i+50, len(toEmails))
		batchEmails := toEmails[i:end]
		batchReqs := make([]*resend.SendEmailRequest, len(batchEmails))
		for j, email := range batchEmails {
			dataIdx := i + j
			data := dataList[0]
			if dataIdx < len(dataList) {
				data = dataList[dataIdx]
			}
			batchReqs = append(batchReqs, &resend.SendEmailRequest{
				To:       []string{email},
				Template: &resend.EmailTemplate{Id: templateID, Variables: data},
			})
		}
		result, err := c.client.Batch.Send(batchReqs)
		if err != nil {
			failureCount += len(batchEmails)
		} else {
			successCount += len(result.Data)
			failureCount += len(result.Errors)
		}
	}
	return map[string]int{"success": successCount, "failure": failureCount}
}
