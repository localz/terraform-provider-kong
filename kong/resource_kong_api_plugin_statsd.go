package kong

import (
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

// Plugin : Kong API statsd plugin request object structure
type StatsdPlugin struct {
	Id                    string    `json:"id,omitempty"`
	ApiId                 string    `json:"-"`
	ConsumerId            string    `json:"consumer_id,omitempty"`
	Name                  string    `json:"name"`
	Config struct {
		Host              string    `json:"host,omitempty"`
		Port              string    `json:"host,omitempty"`
		Prefix            string    `json:"host,omitempty"`
		Metrics 		  []Metrics	`json:"metrics,omitempty"`
	}                               `json:"config,omitempty"`
}

type Metrics []struct {
	Name          string `json:"name,omitempty"`
	SampleRate    int    `json:"sample_rate,omitempty"`
	StatType      string `json:"stat_type,omitempty"`
}

func resourceKongStatsdPlugin() *schema.Resource {
	return &schema.Resource{
		Create: resourceKongStatsdPluginCreate,
		Read:   resourceKongStatsdPluginRead,
		Update: resourceKongStatsdPluginUpdate,
		Delete: resourceKongStatsdPluginDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"consumer_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "The id of the consumer to scope this plugin to.",
			},

			"name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "The name of the plugin to use.",
			},

			"config": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
				Default:  nil,
			},

			"metrics": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"sample_rate": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"stat_type": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				Default:  nil,
			},

			"api": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
		},
	}
}

func resourceKongStatsdPluginCreate(d *schema.ResourceData, meta interface{}) error {
	log.Println("[DEBUG] STEVE resourceKongStatsdPluginCreate ", *d)

	sling := meta.(*sling.Sling)
	plugin := getStatsdPluginFromResourceData(d)

	createdPlugin := plugin

	request := sling.New().BodyJSON(*plugin)
	if plugin.ApiId != "" {
		request = request.Path("apis/").Path(plugin.ApiId + "plugins/")
	}

	log.Println("[DEBUG] STEVE request", request)
	response, error := request.Post("plugins/").ReceiveSuccess(createdPlugin)
	
	if error != nil {
		return fmt.Errorf("error while creating plugin: " + error.Error())
	}

	if response.StatusCode == http.StatusConflict {
		return fmt.Errorf("409 Conflict - use terraform import to manage this plugin.")
	} else if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code received: " + response.Status)
	}

	createdPlugin.Config = plugin.Config

	setStatsdPluginToResourceData(d, createdPlugin)

	return nil
}

func resourceKongStatsdPluginRead(d *schema.ResourceData, meta interface{}) error {
	sling := meta.(*sling.Sling)

	plugin := getPluginFromResourceData(d)

	configuration := make(map[string]interface{})
	for key, value := range plugin.Configuration {
		configuration[key] = value
	}

	response, error := sling.New().Path("plugins/").Get(plugin.ID).ReceiveSuccess(plugin)
	if error != nil {
		return fmt.Errorf("error while updating plugin: " + error.Error())
	}

	if response.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code received: " + response.Status)
	}

	plugin.Configuration = configuration

	setPluginToResourceData(d, plugin)

	return nil
}

func resourceKongStatsdPluginUpdate(d *schema.ResourceData, meta interface{}) error {
	sling := meta.(*sling.Sling)

	plugin := getPluginFromResourceData(d)

	updatedPlugin := getPluginFromResourceData(d)

	response, error := sling.New().BodyJSON(plugin).Path("plugins/").Patch(plugin.ID).ReceiveSuccess(updatedPlugin)
	if error != nil {
		return fmt.Errorf("error while updating plugin: " + error.Error())
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code received: " + response.Status)
	}

	updatedPlugin.Configuration = plugin.Configuration

	setPluginToResourceData(d, updatedPlugin)

	return nil
}

func resourceKongStatsdPluginDelete(d *schema.ResourceData, meta interface{}) error {
	sling := meta.(*sling.Sling)

	plugin := getPluginFromResourceData(d)

	response, error := sling.New().Path("plugins/").Delete(plugin.ID).ReceiveSuccess(nil)
	if error != nil {
		return fmt.Errorf("error while deleting plugin: " + error.Error())
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code received: " + response.Status)
	}

	return nil
}

func getStatsdPluginFromResourceData(d *schema.ResourceData) *StatsdPlugin {
	log.Println("[DEBUG] getStatsdPluginFromResourceData: ", *d)

	plugin := &StatsdPlugin{
		Name:          		d.Get("name").(string),
		ApiId:          	d.Get("api").(string),
		ConsumerId:      	d.Get("consumer_id").(string),
	}

	plugin.Config.Host = d.Get("config.host").(string)
	plugin.Config.Port = d.Get("config.port").(string)
	//plugin.Config.Metrics = d.Get("config.metrics").([]Metrics)

	if id, ok := d.GetOk("id"); ok {
		plugin.Id = id.(string)
	}

	return plugin
}

func setStatsdPluginToResourceData(d *schema.ResourceData, statsdPlugin *StatsdPlugin) {
	d.SetId(statsdPlugin.Id)
	d.Set("name", statsdPlugin.Name)
	d.Set("api", statsdPlugin.ApiId)
	d.Set("consumer_id", statsdPlugin.ConsumerId)

	d.Set("config.host", statsdPlugin.Config.Host)
	d.Set("config.port", statsdPlugin.Config.Port)
	d.Set("metrics", statsdPlugin.Config.Metrics)
}
