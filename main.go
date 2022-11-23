package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/devops"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/resourcemanager"

	jwt "github.com/golang-jwt/jwt/v4"
)

var mySigningKey = []byte("superkeymode1039JnAmm")
var stack_id = os.Getenv("ENV_RM_STACK_ID")
var devops_build_id = os.Getenv("ENV_DEVOPS_BUILD_ID")

func initializeConfigurationProvider() common.ConfigurationProvider {
	privateKey_stage := os.Getenv("ENV_PEM") // set content of PEM in environment variable pem using export pem=$(cat ./ppk)  with ppk a text file that contains the PEM private key00
	privateKey, err := base64.StdEncoding.DecodeString(privateKey_stage)
	if err != nil {
		panic(err)
	}
	tenancyOCID := os.Getenv("ENV_TENANCY_OCID")
	userOCID := os.Getenv("ENV_USER_OCID")
	region := os.Getenv("ENV_REGION")
	fingerprint := os.Getenv("ENV_FINGERPRINT")

	configurationProvider := common.NewRawConfigurationProvider(tenancyOCID, userOCID, region, fingerprint, string(privateKey), nil)
	return configurationProvider

}

// fatalIfError is equivalent to Println() followed by a call to os.Exit(1) if error is not nil
func fatalIfError(err error) {
	if err != nil {
		//log.Fatalln(err.Error())
		log.Printf(err.Error())
	}
}

func getResponseStatusCode(response *http.Response) int {
	return response.StatusCode

}

func RunApplyStackOCI() {
	configurationProvider := initializeConfigurationProvider()
	client, err := resourcemanager.NewResourceManagerClientWithConfigurationProvider(configurationProvider)
	helpers.FatalIfError(err)

	req := resourcemanager.CreateJobRequest{CreateJobDetails: resourcemanager.CreateJobDetails{Operation: resourcemanager.JobOperationApply,
		StackId:             common.String(stack_id),
		JobOperationDetails: resourcemanager.CreateApplyJobOperationDetails{ExecutionPlanStrategy: "AUTO_APPROVED"}}}

	// Send the request using the service client
	resp, err := client.CreateJob(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	fmt.Println(resp)

}

func RunDestroyStackOCI() {
	configurationProvider := initializeConfigurationProvider()
	client, err := resourcemanager.NewResourceManagerClientWithConfigurationProvider(configurationProvider)
	helpers.FatalIfError(err)

	req := resourcemanager.CreateJobRequest{CreateJobDetails: resourcemanager.CreateJobDetails{Operation: resourcemanager.JobOperationDestroy,
		StackId:             common.String(stack_id),
		JobOperationDetails: resourcemanager.CreateDestroyJobOperationDetails{ExecutionPlanStrategy: "AUTO_APPROVED"}}}

	// Send the request using the service client
	resp, err := client.CreateJob(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	fmt.Println(resp)

}

func RunBuildPipeline() {
	configurationProvider := initializeConfigurationProvider()

	client, err := devops.NewDevopsClientWithConfigurationProvider(configurationProvider)
	helpers.FatalIfError(err)

	req := devops.CreateBuildRunRequest{CreateBuildRunDetails: devops.CreateBuildRunDetails{BuildPipelineId: common.String(devops_build_id)}}

	resp, err := client.CreateBuildRun(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	fmt.Println(resp)

}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
	log.Printf("Endpoint Hit: homePage authozied\n")

}

func blankPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, string("Welcome to OCI blank screen, use /oci to create your resources! == VERSION 1.0"))
	log.Printf("Endpoint Hit: blankpage\n")

}

func runOCI(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, string("Running job on STACK, please check the OCI Console for job status!"))
	log.Printf("Running job on STACK...\n")
	RunApplyStackOCI()
}

func runOCIDestroy(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, string("Running DESTROY job on STACK, please check the OCI Console for job status!"))
	log.Printf("Running job on STACK...\n")
	RunDestroyStackOCI()
}

func runPipeline(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, string("Starting build pipeline, please check the OCI Devops Console for job status!"))
	log.Printf("Running build pipeline on OCI Devops...\n")
	RunBuildPipeline()
}

func DestroyAirflow(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("kubectl", "delete", "ns", "airflow")
	out, err := cmd.Output()
	if err != nil {
		log.Printf(err.Error())
	}
	fmt.Println("Output: ", string(out))
}

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Token"] != nil {

			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return mySigningKey, nil
			})

			if err != nil {
				fmt.Fprintf(w, err.Error())
			}

			if token.Valid {
				endpoint(w, r)
			}
		} else {

			fmt.Fprintf(w, "Not Authorized")
			log.Printf("Not Authorized\n")
		}
	})
}

func handleRequests() {
	log.Printf("Starting app...\n")
	http.Handle("/home", isAuthorized(homePage))
	http.HandleFunc("/", blankPage)
	http.Handle("/oci-init", isAuthorized(runOCI))
	http.Handle("/oci-destroy", isAuthorized(runOCIDestroy))
	http.Handle("/airflow-install", isAuthorized(runPipeline))
	http.Handle("/airflow-delete", isAuthorized(DestroyAirflow))
	log.Printf("APP Started, listening on port 8080...\n")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func main() {
	handleRequests()
}
