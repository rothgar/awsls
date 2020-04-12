// Code is generated. DO NOT EDIT.

package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ses"
)

func ListSesActiveReceiptRuleSet(client *Client) error {
	req := client.sesconn.ListReceiptRuleSetsRequest(&ses.ListReceiptRuleSetsInput{})

	resp, err := req.Send(context.Background())
	if err != nil {
		return err
	}

	if len(resp.RuleSets) > 0 {
		for _, r := range resp.RuleSets {
			fmt.Println(*r.Name)
		}
	}

	return nil
}