package deployengine

type K8sDeploymentRequest struct {
	// Application the application that the deployment is for
	Application string `json:"application"`
	// Account the account name that you registered via an agent that you want to deploy in
	Account string `json:"account"`
	// Namespace override the namespaces defined in your manifests
	Namespace string `json:"namespace"`
	// Manifests the array of manifests that you want deployed as part of the deployment
	// There must be one and only one Deployment manifest in this list.
	Manifests []*K8sManifest `json:"manifests"`
	// Define this if you want to do a canary rollout
	CanaryStrategy *K8sCanaryStrategy `json:"canary,omitempty"`
}

type K8sManifestInlineValue struct {
	// Value the raw json or yaml string of the manifest
	Value string `json:"value"`
}

type K8sManifest struct {
	// Name a human-readable name for what the manifest represents
	Name string `json:"name"`
	// InlineValue Define this to supply the manifest inline as part of the request
	InlineValue *K8sManifestInlineValue `json:"inline"`
}

type K8sCanaryStrategy struct {
	// Steps the array of canary steps, this describes how your deployment will be progressively rolled out.
	Steps []*K8sCanaryStep `json:"steps"`
}

// K8sCanaryStep Define one and only one property on this struct!
type K8sCanaryStep struct {
	// SetWeightStep Define this to configure a step that will scale up the canary to a defined weight
	SetWeightStep *SetWeightStep `json:"setWeight,omitempty"`
	// PauseStep Define this to configure a timed pause or a manual judgement
	PauseStep *PauseStep `json:"pause,omitempty"`
}

type SetWeightStep struct {
	// Weight the percentage of weight that you want going to the canary
	Weight int `json:"weight"`
}

type PauseStep struct {
	// Duration the amount of time to wait you must also define Unit
	Duration int `json:"duration,omitempty"`
	// Unit the unit ex: seconds, minutes, hours, days that the Duration int represents
	Unit string `json:"unit,omitempty"`
	// UntilApproved set this to true and omit Duration and Unit to create a manual judgement step
	UntilApproved bool `json:"untilApproved,omitempty"`
}