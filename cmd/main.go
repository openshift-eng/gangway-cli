package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/openshift-eng/gangway-cli/pkg/api"
)

const strFmt = "%-3s | %-38s | %-80s\n"
const maxJobs = 20

// JobRunIdentifier is mirrored to https://github.com/openshift/ci-tools/blob/91aeed7425cb6738b6e36e7f511787b55c9e5267/pkg/jobrunaggregator/jobrunaggregatorlib/util.go#L43
// the output of this command can be fed into ci-tools aggregation commands for further analysis
type JobRunIdentifier struct {
	JobName  string
	JobRunID string
}

type JobStatus struct {
	BuildID string
	JobURL  string
	JobID   string
}

var opts struct {
	initial      string
	latest       string
	envVariables []string
	jobName      string
	apiURL       string
	num          int
	jobsFilePath string
}

var backoff = wait.Backoff{
	Duration: 10 * time.Second,
	Jitter:   0,
	Factor:   2,
	Steps:    100,
}

var cmd = &cobra.Command{
	Use:   "mycli",
	Short: "CLI tool to call an API",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the MY_APPCI_TOKEN environment variable
		appCIToken := os.Getenv("MY_APPCI_TOKEN")
		if appCIToken == "" {
			cmd.Usage() // nolint
			return fmt.Errorf("cluster token required; please set the MY_APPCI_TOKEN variable")
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

		// Add environment variables
		for _, envVar := range opts.envVariables {
			splitVar := strings.SplitN(envVar, "=", 2)
			if len(splitVar) != 2 {
				cmd.Usage() // nolint
				return fmt.Errorf("invalid environment variable, should be in format of VAR=VALUE: %q", envVar)
			}
			spec.PodSpecOptions.Envs[splitVar[0]] = splitVar[1]
		}

		// Convert image spec to JSON and pretty print in case someone needs to
		// debug it.
		data, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			return err
		}

		if opts.num > maxJobs {
			return fmt.Errorf("aborting since %d exceeds max value of --n which is %d", opts.num, maxJobs)
		}

		var jobRunIdentifiers = make([]JobRunIdentifier, 0)

		fmt.Println(string(data))

		fmt.Printf(strFmt, "Job", "ID", "URL")
		fmt.Println("---------------------------------------------------------------------------")

		for i := 0; i < opts.num; i++ {
			// Make the HTTP request
			resp, err := launchJob(appCIToken, opts.apiURL, data)
			if err != nil {
				return err
			}

			var jobInfo struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&jobInfo); err != nil {
				resp.Body.Close() // nolint
				return err
			}
			resp.Body.Close() // nolint

			// Get the job URL from prow easy access
			jobStatus := JobStatus{JobID: jobInfo.ID}
			err = wait.ExponentialBackoff(backoff, func() (done bool, err error) {
				err2 := jobStatus.getJobURL()
				if err2 != nil {
					return false, err2
				}
				if len(jobStatus.JobURL) == 0 {
					fmt.Println("Empty job URL will attempt retry")
					return false, nil
				}
				return true, nil
			})

			if err != nil {
				return err
			}

			jobRunIdentifier := JobRunIdentifier{
				JobName:  opts.jobName,
				JobRunID: jobStatus.BuildID,
			}
			jobRunIdentifiers = append(jobRunIdentifiers, jobRunIdentifier)

			// Print job info in tabular format
			fmt.Printf(strFmt, strconv.Itoa(i+1), jobInfo.ID, jobStatus.JobURL)

			// Sleep to avoid hitting the api too hard
			time.Sleep(time.Second)
		}

		outputJobRunIdentifiers(opts.jobsFilePath, jobRunIdentifiers)

		return nil
	},
}

func NewCommand() *cobra.Command {
	cmd.Flags().StringVarP(&opts.initial, "initial", "i", "", "Initial image")
	cmd.Flags().StringVarP(&opts.latest, "latest", "l", "", "Latest image")
	cmd.Flags().StringArrayVarP(&opts.envVariables, "env", "e", nil, "Environment variables, VAR=VALUE")
	cmd.Flags().StringVarP(&opts.jobName, "job-name", "j", "", "Job name")
	cmd.Flags().StringVarP(&opts.apiURL, "api-url", "u", "", "API URL")
	cmd.Flags().IntVarP(&opts.num, "num-jobs", "n", 1, fmt.Sprintf("Number of times to launch the job (max is %d)", maxJobs))
	cmd.Flags().StringVarP(&opts.jobsFilePath, "jobs-file-path", "p", "", "Save JobRunIdentifier JSON in the specified file path")

	return cmd
}

// launchJob launches a prow job using the gangway api authenticated with the token.
func launchJob(appCIToken, apiURL string, data []byte) (*http.Response, error) {
	url := apiURL + "/v1/executions/" + opts.jobName
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+appCIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request to gangway api: %v", err)
	}

	return resp, nil
}

// getJobURL gets the url from prow so the user has a place to browse to see the
// status of the prow job.  Prow does not immediately have the prow job so we wait.
func (j *JobStatus) getJobURL() error {
	url := "https://prow.ci.openshift.org/prowjob?prowjob=" + j.JobID
	var resp *http.Response
	var err error

	resp, err = http.Get(url)
	if err != nil {
		fmt.Printf("Error getting job URL: %v\n", err)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return fmt.Errorf("error reading response body: %v", err)
	}
	resp.Body.Close()

	// Search YAML document parts to find the section with status.url
	documents := strings.Split(string(body), "---")
	for _, doc := range documents {
		var jobInfo map[string]interface{}
		if err := yaml.Unmarshal([]byte(doc), &jobInfo); err != nil {
			continue
		}
		if status, ok := jobInfo["status"].(map[interface{}]interface{}); ok {

			// try to get the build_id as well
			// but only return when we have a valid URL
			if bid, ok := status["build_id"]; ok {
				j.BuildID = bid.(string)
			}

			if url, ok := status["url"]; ok {
				j.JobURL = url.(string)
				return nil
			}
		}
	}

	return nil
}

func outputJobRunIdentifiers(jobsFilePath string, jobRunIdentifiers []JobRunIdentifier) {

	fileName := fmt.Sprintf("gangway_%s_%s.json", opts.jobName, time.Now().Format(time.RFC3339Nano))
	output, err := json.Marshal(jobRunIdentifiers)
	if err != nil {
		fmt.Printf("Failed to marshal JSON for JobRunIdentifiers: %v", err)
	} else {

		// check to see if we end with path separator
		// if so set it to empty string
		delimiter := string(os.PathSeparator)
		if len(jobsFilePath) > 0 {
			if strings.HasSuffix(jobsFilePath, delimiter) {
				delimiter = ""
			}
			err := os.MkdirAll(jobsFilePath, os.ModePerm)
			if err != nil {
				fmt.Printf("Error creating directory: %v", err)
			} else {
				err := os.WriteFile(fmt.Sprintf("%s%s%s", jobsFilePath, delimiter, fileName), output, os.ModePerm)
				if err != nil {
					fmt.Printf("Error writing file: %v", err)
				}
			}
		}
	}

}
