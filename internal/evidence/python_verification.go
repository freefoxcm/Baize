package evidence

// pythonSegmentIsVerification accepts isolation flags before a trusted test
// runner while keeping inline source and arbitrary modules untrusted.
func pythonSegmentIsVerification(args []string) bool {
	for len(args) > 0 && (args[0] == "-B" || args[0] == "-E") {
		args = args[1:]
	}
	return len(args) > 1 && args[0] == "-m" && hasCommandArg(args[1:2], "pytest", "unittest")
}
