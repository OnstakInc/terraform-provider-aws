package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEcrRepository() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcrRepositoryRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsEcrRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	params := &ecr.DescribeRepositoriesInput{
		RepositoryNames: aws.StringSlice([]string{name}),
	}
	log.Printf("[DEBUG] Reading ECR repository: %s", params)
	out, err := conn.DescribeRepositories(params)
	if err != nil {
		if isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") {
			return fmt.Errorf("ECR Repository (%s) not found", name)
		}
		return fmt.Errorf("error reading ECR repository: %w", err)
	}

	repository := out.Repositories[0]
	arn := aws.StringValue(repository.RepositoryArn)

	d.SetId(aws.StringValue(repository.RepositoryName))
	d.Set("arn", arn)
	d.Set("registry_id", repository.RegistryId)
	d.Set("name", repository.RepositoryName)
	d.Set("repository_url", repository.RepositoryUri)

	tags, err := keyvaluetags.EcrListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for ECR Repository (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
