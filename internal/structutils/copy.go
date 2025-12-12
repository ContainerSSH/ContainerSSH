package structutils

import "github.com/jinzhu/copier"

// Copy deep copies a struct pointer from source to destination.
func Copy(destination interface{}, source interface{}) error {
	err := copier.CopyWithOption(destination, source, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return err
}
