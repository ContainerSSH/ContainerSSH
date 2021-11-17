package message

// LabelName is a name for a Message label. Can only contain A-Z, a-z, 0-9, -, _.
type LabelName string

// LabelValue is a string, int, bool, or float.
type LabelValue interface{}

// Labels is a map linking together a label name with a label value.
type Labels map[LabelName]LabelValue
