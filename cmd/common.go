package cmd

var (
	profile      string
	skipSteps    = []string{}
	outputFormat string
	Config       string
)

func mapFromSlice(slice []string) map[string]string {
	resultantMap := make(map[string]string)
	for _, step := range slice {
		resultantMap[step] = ""
	}
	return resultantMap
}
