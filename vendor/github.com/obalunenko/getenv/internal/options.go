package internal

// Parameters is a struct for holding parameters for the parser.
// It is used to pass parameters to the parser.
// Separator is a separator for the environment variable that holds slice.
// Layout is a layout for the time.Time.
type Parameters struct {
	Separator string
	Layout    string
}
