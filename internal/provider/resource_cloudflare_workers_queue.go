package provider

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceCloudflareWorkersQueue() *schema.Resource {
	return &schema.Resource{
		Schema:        resourceCloudflareWorkersQueueSchema(),
		CreateContext: resourceCloudflareWorkersQueueCreate,
		ReadContext:   resourceCloudflareWorkersQueueRead,
		UpdateContext: resourceCloudflareWorkersQueueUpdate,
		DeleteContext: resourceCloudflareWorkersQueueDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceCloudflareWorkersQueueImport,
		},
		Description: "Provides the ability to manage Cloudflare Workers Queue features.",
	}
}

func resourceCloudflareWorkersQueueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflare.API)

	accountID := d.Get("account_id").(string)
	if accountID == "" {
		accountID = client.AccountID
	}

	queueName := d.Get("name").(string)
	req := cloudflare.CreateQueueParams{
		Name: queueName,
	}

	tflog.Debug(ctx, fmt.Sprintf("[Info] Creating Cloudflare Workers Queue from struct: %+v", req))

	r, err := client.CreateQueue(ctx, cloudflare.AccountIdentifier(accountID), req)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "error creating workers queue"))
	}

	if r.ID == "" {
		return diag.FromErr(fmt.Errorf("failed to find id in Create response; resource was empty"))
	}

	d.SetId(r.ID)

	tflog.Info(ctx, fmt.Sprintf("Cloudflare Workers Queue ID: %s. Name: %s", d.Id(), queueName))

	return resourceCloudflareWorkersQueueRead(ctx, d, meta)
}

func resourceCloudflareWorkersQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflare.API)
	queueID := d.Id()

	accountID := d.Get("account_id").(string)
	if accountID == "" {
		accountID = client.AccountID
	}

	resp, _, err := client.ListQueues(ctx, cloudflare.AccountIdentifier(accountID), cloudflare.ListQueuesParams{})
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "error reading workers "))
	}

	var queue cloudflare.Queue
	for _, r := range resp {
		if r.ID == queueID {
			queue = r
			break
		}
	}

	if queue.ID == "" {
		d.SetId("")
		return nil
	}

	d.Set("account_id", accountID)

	return nil
}

func resourceCloudflareWorkersQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflare.API)

	accountID := d.Get("account_id").(string)
	if accountID == "" {
		accountID = client.AccountID
	}

	// TODO(soon) fix the cloudflare-go UpdateQueue implementation: updating a queue should accept the existing name, as well as the new name. Other properties are read-only.
	_, err := client.UpdateQueue(ctx, cloudflare.AccountIdentifier(accountID), cloudflare.UpdateQueueParams{
		Name: d.Get("name").(string),
	})
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "error updating workers queue"))
	}

	return resourceCloudflareWorkersQueueRead(ctx, d, meta)
}

func resourceCloudflareWorkersQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflare.API)
	accountID := d.Get("account_id").(string)
	if accountID == "" {
		accountID = client.AccountID
	}

	tflog.Info(ctx, fmt.Sprintf("Deleting Cloudflare Workers Queue with id: %+v", d.Id()))

	err := client.DeleteQueue(ctx, cloudflare.AccountIdentifier(accountID), d.Get("name").(string))
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "error deleting workers queue"))
	}

	d.SetId("")
	return nil
}

// TODO(now) is this needed?
func resourceCloudflareWorkersQueueImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
