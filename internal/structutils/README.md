<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Struct Manipulation Library</h1>

This library provides methods for manipulating structs.

## Copying structures

The `structutils.Copy()` method provides a deep-copy mechanism for structs:

```go
oldData := yourStruct{}
newData := yourStruct{}
err := structutils.Copy(&newData, &oldData)
```

The `newData` will be completely independent from `oldData`, including the copying of any pointers.

## Merging structures

The `structutils.Merge()` method provides a deep merge of two structs:

```go
oldData := yourStruct{}
newData := yourStruct{}
err := structutils.Merge(newData, oldData)
```

The `Merge` method will copy non-default values from `oldData` to `newData`.

## Adding default values

The `structutils.Defaults()` method loads the default values from the `default` tag in a struct:

```go
type yourStruct struct {
	Text string `default:"Hello world!"`
	Decision bool `default:"true"`
	Number int `default:"42"`
}

//...

func main() {
    data := yourStruct{}
    structutils.Defaults(&data)
    // testdata will now contain the default values
}
```

Default values can be provided as follows:

- Scalars can be provided directly.
- Maps, structs, etc. can be provided in JSON format.
- `time.Duration` can be provided in text format (e.g. 60s).
- Structs may implement a receiver with the method called SetDefaults() as described in [defaults.go](defaults.go).
