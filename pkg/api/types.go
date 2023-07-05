package api

// PodSpecOptions struct represents the options specific to the pod spec.
type PodSpecOptions struct {
	Envs map[string]string `json:"envs"`
}

// ImageSpec struct represents the image specification details.
type ImageSpec struct {
	JobExecutionType string         `json:"job_execution_type"`
	PodSpecOptions   PodSpecOptions `json:"pod_spec_options"`
}
