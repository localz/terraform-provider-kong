package kong

import (
	"fmt"
	"log"
	"net/http"

	"github.com/davecgh/go-spew/spew"

	"github.com/dghubble/sling"
	"github.com/hashicorp/terraform/helper/schema"
)

type JSONConfig struct {
	Add struct {
		Headers     []interface{} `json:"headers,omitempty"`
		Querystring []interface{} `json:"querystring,omitempty"`
		Body        []interface{} `json:"body,omitempty"`
	}
}

type Plugin struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// Configuration map[string]interface{} `json:"config,omitempty"`
	// Configuration map[string]JSONConfig `json:"config,omitempty"`
	Configuration interface{} `json:"config,omitempty"`
	API           string      `json:"-"`
}

// var mods map[string]*schema.Schema = map[string]*schema.Schema{
// 	"header": &schema.Schema{
// 		Type:     schema.TypeList,
// 		Elem:     schema.TypeString,
// 		Optional: true,
// 	},
// }
//
// var options map[string]*schema.Schema = map[string]*schema.Schema{
// 	"add": &schema.Schema{
// 		Type:     schema.TypeList,
// 		Elem:     mods,
// 		Optional: true,
// 	},
// "add": &schema.Schema{
// 	Type:     schema.TypeList,
// 	Elem:     mods,
// 	Optional: true,
// },
// }

func resourceKongPlugin() *schema.Resource {
	return &schema.Resource{
		Create: resourceKongPluginCreate,
		Read:   resourceKongPluginRead,
		Update: resourceKongPluginUpdate,
		Delete: resourceKongPluginDelete,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"add": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"headers": &schema.Schema{
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
				Default: nil,
			},

			"api": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceKongPluginCreate(d *schema.ResourceData, meta interface{}) error {
	sling := meta.(*sling.Sling)

	plugin := getPluginFromResourceData(d)

	createdPlugin := getPluginFromResourceData(d)

	response, error := sling.New().BodyJSON(plugin).Path("apis/").Path(plugin.API + "/").Post("plugins/").ReceiveSuccess(createdPlugin)
	str := spew.Sdump(response)
	log.Println("KONG API PLUGIN")
	log.Println(str)

	log.Println(spew.Sdump(error))
	if error != nil {
		return fmt.Errorf("Error while creating plugin.")
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf(response.Status)
	}

	setPluginToResourceData(d, createdPlugin)

	return nil
}

func resourceKongPluginRead(d *schema.ResourceData, meta interface{}) error {
	sling := meta.(*sling.Sling)

	plugin := getPluginFromResourceData(d)

	response, error := sling.New().Path("apis/").Path(plugin.API + "/").Path("plugins/").Get(plugin.ID).ReceiveSuccess(plugin)
	if error != nil {
		return fmt.Errorf("Error while updating plugin.")
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(response.Status)
	}

	setPluginToResourceData(d, plugin)

	return nil
}

func resourceKongPluginUpdate(d *schema.ResourceData, meta interface{}) error {
	sling := meta.(*sling.Sling)

	plugin := getPluginFromResourceData(d)

	updatedPlugin := getPluginFromResourceData(d)

	response, error := sling.New().BodyJSON(plugin).Path("apis/").Path(plugin.API + "/").Path("plugins/").Patch(plugin.ID).ReceiveSuccess(updatedPlugin)
	if error != nil {
		return fmt.Errorf("Error while updating plugin.")
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(response.Status)
	}

	setPluginToResourceData(d, updatedPlugin)

	return nil
}

func resourceKongPluginDelete(d *schema.ResourceData, meta interface{}) error {
	sling := meta.(*sling.Sling)

	plugin := getPluginFromResourceData(d)

	response, error := sling.New().Path("apis/").Path(plugin.API + "/").Path("plugins/").Delete(plugin.ID).ReceiveSuccess(nil)
	if error != nil {
		return fmt.Errorf("Error while deleting plugin.")
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf(response.Status)
	}

	return nil
}

func getPluginFromResourceData(d *schema.ResourceData) *Plugin {
	log.Print("GET PLUGIN")
	log.Print(spew.Sdump(d.Get("config")))

	plugin := &Plugin{
		Name:          d.Get("name").(string),
		Configuration: d.Get("config").(interface{}),
		API:           d.Get("api").(string),
	}

	if id, ok := d.GetOk("id"); ok {
		plugin.ID = id.(string)
	}

	return plugin
}

func setPluginToResourceData(d *schema.ResourceData, plugin *Plugin) {
	log.Print("SET PLUGIN")
	log.Print(spew.Sdump(d))
	d.SetId(plugin.ID)
	d.Set("name", plugin.Name)
	d.Set("config", plugin.Configuration)
	d.Set("api", plugin.API)
}
