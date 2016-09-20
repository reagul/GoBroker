package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"net/http"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/cloudfoundry-community/types-cf"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/kr/pretty"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

type serviceInstanceResponse struct {
	DashboardURL string `json:"dashboard_url"`
}

type serviceBindingResponse struct {
	Credentials    map[string]interface{} `json:"credentials"`
	SyslogDrainURL string                 `json:"syslog_drain_url,omitempty"`
}

type ServiceCreationRequest struct {
	InstanceID       string            `json:"-"`
	ServiceID        string            `json:"service_id"`
	PlanID           string            `json:"plan_id"`
	OrganizationGUID string            `json:"organization_guid"`
	SpaceGUID        string            `json:"space_guid"`
	Parameters       map[string]string `json:parameters`
}

var solrConfigSetName, solrEndPoint, serviceName, servicePlan, baseGUID, authUser, authPassword, tags, imageURL string
var defaultServiceBinding serviceBindingResponse
var credentials string

var appURL              string
var parameterMap        map[string]string
var instanceCollections map[string]string
var collectionName      string
var serviceCreationRequest ServiceCreationRequest

func brokerCatalog() (int, []byte) {
	tagArray := []string{}
	if len(tags) > 0 {
		tagArray = strings.Split(tags, ",")
	}
	var requires []string
	if len(defaultServiceBinding.SyslogDrainURL) > 0 {
		requires = []string{"syslog_drain"}
	}
	catalog := cf.Catalog{
		Services: []*cf.Service{
			{
				ID:          baseGUID + "-service-" + serviceName,
				Name:        serviceName,
				Description: "Shared service for " + serviceName,
				Bindable:    true,
				Tags:        tagArray,
				Requires:    requires,
				Metadata: &cf.ServiceMeta{
					DisplayName: serviceName,
					ImageURL:    imageURL,
				},
				Plans: []*cf.Plan{
					{
						ID:          baseGUID + "-plan-" + servicePlan,
						Name:        servicePlan,
						Description: "Shared service for " + serviceName,
						Free:        true,
					},
				},
			},
		},
	}
	json, err := json.Marshal(catalog)
	if err != nil {
		fmt.Println("Um, how did we fail to marshal this catalog:")
		fmt.Printf("%# v\n", pretty.Formatter(catalog))
		return 500, []byte{}
	}
	return 200, json
}

func createServiceInstance(r *http.Request, params martini.Params) (int, []byte) {
	serviceID := params["service_id"]

	
    decoder := json.NewDecoder(r.Body) 

    err := decoder.Decode(&serviceCreationRequest)
	
	


		fmt.Println("what cloud controller passes to us: ")
		fmt.Printf("%# v\n", pretty.Formatter(serviceCreationRequest))
		//return 500, []byte{}
	

	collectionName = serviceCreationRequest.Parameters["collection"]
	fmt.Printf("Collection is %s for Request %s \n", collectionName, r,  )
	//err := json.UnMarshal(params["parameters"], parameterMap)
	//collectionName = parameterMap["collection"]
	//solrEndPoint  = defaultServiceBinding.Credentials["SolrEndpoint"].(string)
	fmt.Printf("Creating service instance %s for service %s plan %s\n", serviceID, serviceName, servicePlan, )

	instance := serviceInstanceResponse{DashboardURL: fmt.Sprintf("%s/dashboard", appURL)}
	json, err := json.Marshal(instance)
	if err != nil {
		fmt.Println("Um, how did we fail to marshal this service instance:")
		fmt.Printf("%# v\n", pretty.Formatter(instance))
		return 500, []byte{}
	}
    

	  //  res, err := http.Get(solrEndPoint + "/admin/collections?action=CREATE&name=" + collectionName + "&numShards=2&replicationFactor=2&maxShardsPerNode=4&collection.configName=" + solrConfigSetName)

    if err != nil {
        log.Fatal(err)
    }

   // createCollection, err := ioutil.ReadAll(res.Body)

    //res.Body.Close()
    if err != nil {
    	log.Fatal(err)
    }
   // fmt.Printf("%s", createCollection)
    instanceCollections[serviceID] = collectionName
	return 201, json
}

func deleteServiceInstance(params martini.Params) (int, string) {
	serviceID := params["serviceID"]
	fmt.Printf("Deleting service instance %s for service %s plan %s\n", serviceID, serviceName, servicePlan)
	return 200, "{}"
}

func createServiceBinding(params martini.Params) (int, []byte) {
	var serviceBinding serviceBindingResponse

	serviceID := params["service_id"]
	serviceBindingID := params["binding_id"]
	fmt.Printf("Creating service binding %s for service %s plan %s instance %s\n",
		serviceBindingID, serviceName, servicePlan, serviceID)

    

    json.Unmarshal([]byte(credentials), &serviceBinding.Credentials)
	//serviceBinding := defaultServiceBinding 
	//delete (serviceBinding.Credentials, "SolrEndpoint");

	fmt.Println("this is the before binding response")
	//json.Unmarshal([]byte(credentials), &serviceBinding.Credentials)
	fmt.Printf("%# v\n", pretty.Formatter(serviceBinding))

	serviceBinding.Credentials["SolrEndpoint"] = serviceBinding.Credentials["SolrEndpoint"].(string) + "/" + instanceCollections[serviceID] 
	
     fmt.Println("this is the after binding response")
	//json.Unmarshal([]byte(credentials), &serviceBinding.Credentials)
	fmt.Printf("%# v\n", pretty.Formatter(serviceBinding))


	json, err := json.Marshal(serviceBinding)

	if err != nil {
		fmt.Println("Um, how did we fail to marshal this binding:")
		fmt.Printf("%# v\n", pretty.Formatter(serviceBinding))
		return 500, []byte{}
	}
	return 201, json
}

func deleteServiceBinding(params martini.Params) (int, string) {
	serviceID := params["service_id"]

	serviceBindingID := params["binding_id"]
	fmt.Printf("Delete service binding %s for service %s plan %s instance %s\n",
		serviceBindingID, serviceName, servicePlan, serviceID)
	return 200, "{}"
}

func showServiceInstanceDashboard(params martini.Params) (int, string) {
	fmt.Printf("Show dashboard for service %s plan %s\n", serviceName, servicePlan)
	return 200, "<a href=\"http://dashboard.app.mgpcf.net/dashboard.html\"> Dashboard </a>"
}

func main() {

    instanceCollections  = make (map[string]string,100)
	m := martini.Classic()

	baseGUID = os.Getenv("BASE_GUID")
	if baseGUID == "" {
		baseGUID = "29140B3F-0E69-4C7E-8A35"
	}
	serviceName = os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "some-service-name" // replace with cfenv.AppName
	}
	servicePlan = os.Getenv("SERVICE_PLAN")
	if servicePlan == "" {
		servicePlan = "shared"
	}

	 solrConfigSetName = os.Getenv("SOLR_CONFIG_SET")
    if solrConfigSetName == "" {
       solrConfigSetName = "logstash_conf"
    }

	authUser = os.Getenv("AUTH_USER")
	authPassword = os.Getenv("AUTH_PASSWORD")
	if (authUser != "") && (authPassword != "") {
		// secure service broker with basic auth if both env variables are set
		m.Use(auth.Basic(authUser, authPassword))
	}

	credentials = os.Getenv("CREDENTIALS")
	if credentials == "" {
		credentials = "{\"port\": \"4000\"}"
	}
	tags = os.Getenv("TAGS")
	imageURL = os.Getenv("IMAGE_URL")

	json.Unmarshal([]byte(credentials), &defaultServiceBinding.Credentials)
	fmt.Printf("%# v\n", pretty.Formatter(defaultServiceBinding))


	defaultServiceBinding.SyslogDrainURL = os.Getenv("SYSLOG_DRAIN_URL")


	appEnv, err := cfenv.Current()
	if err == nil {
		appURL = fmt.Sprintf("https://%s", appEnv.ApplicationURIs[0])
	} else {
		appURL = "http://localhost:5000"
	}
	fmt.Println("Running as", appURL)

	// Cloud Foundry Service API
	m.Get("/v2/catalog", brokerCatalog)
	m.Put("/v2/service_instances/:service_id", createServiceInstance)
	m.Delete("/v2/service_instances/:service_id", deleteServiceInstance)
	m.Put("/v2/service_instances/:service_id/service_bindings/:binding_id", createServiceBinding)
	m.Delete("/v2/service_instances/:service_id/service_bindings/:binding_id", deleteServiceBinding)

	// Service Instance Dashboard
	m.Get("/dashboard", showServiceInstanceDashboard)

	m.Run()
}
