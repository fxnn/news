package main

// Summarize generates a summary for the given text using the default LLM client.
func Summarize(text string) (string, error) {
	// Delegate to the defaultLLM implementation (currently the stub).
	return defaultLLM.Summarize(text)
}
