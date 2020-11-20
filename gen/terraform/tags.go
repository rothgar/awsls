// +build codegen

package terraform

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/apex/log"
	genutil "github.com/jckuester/awsls/gen/util"
	"github.com/jckuester/awsls/util"
)

// GenerateResourceTypesWithTagsList generates code of a list of Terraform resource types that support tags
// and writes the code to directory outputPath.
func GenerateResourceTypesWithTagsList(resourceTypes []string, outputPath string) ([]string, error) {
	awsClientKey := util.AWSClientKey{
		Profile: os.Getenv("AWS_PROFILE"),
		Region:  os.Getenv("AWS_DEFAULT_REGION"),
	}

	providers, err := util.NewProviderPool(
		[]util.AWSClientKey{awsClientKey}, "3.16.0", "~/.awsls", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Terraform AWS provider: %s", err)
	}

	provider, ok := providers[awsClientKey]
	if !ok {
		return nil, fmt.Errorf("Terraform AWS provider not found: %s", err)
	}

	defer func() {
		for _, p := range providers {
			_ = p.Close()
		}
	}()

	var resourceTypesWithTags []string

	for _, rType := range resourceTypes {
		schema, err := provider.GetSchemaForResource(rType)
		if err != nil {
			return nil, fmt.Errorf("failed to get schema for resource: %s", err)
		}

		_, ok := schema.Block.Attributes["tags"]
		if ok {
			resourceTypesWithTags = append(resourceTypesWithTags, rType)
		}
	}

	err = writeResourceTypesWithTags(outputPath, resourceTypesWithTags)
	if err != nil {
		return nil, err
	}

	log.WithField("length", len(resourceTypesWithTags)).Infof("Generated list of Terraform AWS resource types " +
		"that support tags")

	return resourceTypesWithTags, nil
}

func writeResourceTypesWithTags(outputPath string, resourceTypes []string) error {
	err := os.MkdirAll(outputPath, 0775)
	if err != nil {
		return fmt.Errorf("failed to create directory: %s", err)
	}

	code, err := resourceTypesWithTagsGoCode(resourceTypes)
	if err != nil {
		return fmt.Errorf("failed to generate Go code: %s", err)
	}

	err = genutil.WriteGoFile(
		filepath.Join(outputPath, "tags.go"),
		genutil.CodeLayout,
		"",
		"resource",
		code,
	)

	if err != nil {
		return fmt.Errorf("failed to write Go code to file: %s", err)
	}

	return nil
}

func resourceTypesWithTagsGoCode(terraformTypes []string) (string, error) {
	var buf bytes.Buffer
	err := resourceTypesWithTagsTmpl.Execute(&buf, terraformTypes)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

var resourceTypesWithTagsTmpl = template.Must(template.New("resourceHasTags").Parse(`
// TypesWithTags is a list of all resource types that support tags.
var TypesWithTags = []string{
{{range .}}"{{.}}",
{{end}}}
`))
