// Copyright 2024 Google LLC
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
	"encoding/json"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceJDBCDriver() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDBCDriverCreate,
		Read:   resourceLocalArtifactRead,   // Reusing existing read logic
		Delete: resourceLocalArtifactDelete, // Reusing existing delete logic
		Exists: resourceLocalArtifactExists, // Reusing existing existence check

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
				Description: "The local path to the JAR binary for the artifact.",
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

func resourceJDBCDriverCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)
	namespace := d.Get("namespace").(string)
	name := d.Get("name").(string)
	version := d.Get("version").(string)
	jarPath := d.Get("jar_binary_path").(string)

	jarBytes, err := ioutil.ReadFile(jarPath)
	if err != nil {
		return err
	}

	// 2. Build the necessary headers
	headers := map[string]string{
		"Content-Type":     "application/java-archive",
		"Artifact-Version": version,
	}

	if archiveName, ok := d.GetOk("archive_name"); ok {
		headers["x-archive-name"] = archiveName.(string)
	}

	if pluginsJsonStr, ok := d.GetOk("plugins"); ok {
		pluginList := pluginsJsonStr.([]interface{})
		var plugins []map[string]string

		for _, p := range pluginList {
			pMap := p.(map[string]interface{})
			plugin := map[string]string{
				"name":      pMap["name"].(string),
				"type":      pMap["type"].(string),
				"className": pMap["class_name"].(string),
			}

			if desc, ok := pMap["description"]; ok && desc.(string) != "" {
				plugin["description"] = desc.(string)
			}

			plugins = append(plugins, plugin)
		}

		pluginsJSON, err := json.Marshal(plugins)
		if err != nil {
			return err
		}
		headers["artifact-plugins"] = string(pluginsJSON)
	}

	addr := urlJoin(config.host, "/v3/namespaces", namespace, "/artifacts", name)

	if err := uploadPluginJar(config, addr, jarBytes, headers); err != nil {
		return err
	}

	d.SetId(name)
	return nil
}
