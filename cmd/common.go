package cmd

var (
	profile      string
	skipSteps    = []string{}
	outputFormat string
	Config       string
)

func getStepsToSkip() map[string]string {
	skipStepsMap := make(map[string]string)
	for _, step := range skipSteps {
		skipStepsMap[step] = ""
	}
	return skipStepsMap
}
