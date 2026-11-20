# The ContainerSSH Binary Audit Log Format, version {{ .Version }} (draft)

The ContainerSSH audit log is stored in [CBOR](https://cbor.io/) + GZIP format.

However, before GZIP decoding you must provide/decode the **file header**. The file header is {{ .HeaderLength }} bytes long. The fist {{ .MagicLength }} bytes must contain the string `{{ .MagicValue }}` in UTF-8 encoding, the rest padded with 0 bytes. The last 8 bytes contain the audit log format version number as a 64 bit unsigned little endian integer.

```
Header {
    Magic{{ "\t" }}[{{ .MagicLength }}]byte{{ "\t" }}# Must always be {{ .MagicValue }}\0
    Version{{ "\t" }}uint64{{ "\t" }}# Little endian encoding
}
```

After the first {{ .HeaderLength }} bytes you will have to GZIP-decode the rest of the file. The result is a sequence of CBOR-encoded messages, one after the other, with no enclosing array (version 1 wrapped all messages in a single CBOR array). Decode messages until the stream ends. The GZIP stream may lack its end marker, so some GZIP readers report an unexpected end of file, and a log that is still being written may end in the middle of a message. Treat both as the end of the log. Each message is a CBOR map with the following keys:

```
{{ .Message.Name }} {
{{ range .Message.Fields }}{{ with .Comment }}{{ "\t# " }}{{ . }}
{{ end }}{{ "\t" }}{{ .Name }}{{ "\t" }}{{ .DataType }}
{{ end }}}
```

The audit log protocol has the following message types at this time:

| Message type ID | Name | Payload type |
|-----------------|------|--------------|
{{ range $type, $payload := .Payloads }}| {{ $type.Code }} | {{ $type.Name }} | {{ with $payload }}[{{.Name}}](#{{.Name}}){{else}}*none*{{ end }} |
{{ end }}
{{ range $code, $payload := .Payloads }}{{ with $payload }}## {{ .Name }}
{{ with .Description }}
{{ . }}
{{ end }}
```
{{ .Name }} {
{{ range .Fields }}{{ "\t" }}{{ .Name }}{{ "\t" }}{{ .DataType }}{{ with .Comment }}{{ "\t# " }}{{ . }}{{ end }}
{{ end }}}
```

{{ end }}{{ end -}}