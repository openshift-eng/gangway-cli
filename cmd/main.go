package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/spf13/cobra"

	"github.com/stbenjam/gangway-cli/pkg/api"
)

var opts struct {
	initial string
	latest  string
	jobName string
	apiURL  string
}

var cmd = &cobra.Command{
	Use:   "mycli",
	Short: "CLI tool to call an API",
	Run: func(cmd *cobra.Command, args []string) {
		// Get the MY_APPCI_TOKEN environment variable
		appCIToken := os.Getenv("MY_APPCI_TOKEN")
		if appCIToken == "" {
			fmt.Println("Error: Cluster token required.  Please set the MY_APPCI_TOKEN variable.")
			cmd.Usage()
			os.Exit(1)
		}

		// Parse image spec
		spec := api.ImageSpec{
			JobExecutionType: "1",
			PodSpecOptions: api.PodSpecOptions{
				Envs: map[string]string{
					"RELEASE_IMAGE_INITIAL": opts.initial,
					"RELEASE_IMAGE_LATEST":  opts.latest,
				},
			},
		}

		// Convert image spec to JSON
		data, err := json.Marshal(spec)
		if err != nil {
			fmt.Printf("Error converting spec to JSON: %v", err)
			os.Exit(1)
		}

		fmt.Println(string(data))

		// Make the HTTP request
		url := opts.apiURL + "/v1/executions/" + opts.jobName
		fmt.Println(url)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			fmt.Printf("Error creating HTTP request: %v", err)
			os.Exit(1)
		}

		req.Header.Set("Authorization", "Bearer "+appCIToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making HTTP request: %v", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		fmt.Println("HTTP Response:")
		if dump, err := httputil.DumpResponse(resp, true); err != nil {
			fmt.Printf("error: %+v", err)
		} else {
			fmt.Println(string(dump))
		}
	},
}

func NewCommand() *cobra.Command {
	cmd.Flags().StringVarP(&opts.initial, "initial", "i", "", "Initial image")
	cmd.Flags().StringVarP(&opts.latest, "latest", "l", "", "Latest image")
	cmd.Flags().StringVarP(&opts.jobName, "job-name", "j", "", "Job name")
	cmd.Flags().StringVarP(&opts.apiURL, "api-url", "u", "", "API URL")

	return cmd
}
