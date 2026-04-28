// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cdap

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceGCSJDBCDriver supports deploying a JDBC driver artifact by providing a GCS path.
func resourceGCSJDBCDriver() *schema.Resource {
	return &schema.Resource{
		Create: resourceGCSJDBCDriverCreate,
		Read:   resourceLocalArtifactRead,
		Delete: resourceLocalArtifactDelete,
		Exists: resourceLocalArtifactExists,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the artifact.",
			},
			"namespace": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of the namespace in which this resource belongs. If not provided, the default namespace is used.",
				DefaultFunc: func() (interface{}, error) {
					return defaultNamespace, nil
				},
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The version of the artifact.",
			},
			"jar_binary_path": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GCS path (gs://...) to the JAR binary for the artifact.",
			},
			"archive_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The archive name header (e.g., mysql-connector-java).",
			},
			"plugins": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "List of plugins to declare in the artifact headers.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"class_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceGCSJDBCDriverCreate(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()
	config := m.(*Config)
	jarPath := d.Get("jar_binary_path").(string)

	jarBytes, err := readObject(ctx, config.storageClient, jarPath)
	if err != nil {
		return err
	}

	return deployJDBCDriver(d, config, jarBytes)
}
