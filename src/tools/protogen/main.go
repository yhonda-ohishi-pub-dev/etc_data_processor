package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/internal/models"
)

const protoTemplate = `syntax = "proto3";

package {{ .Package }};

option go_package = "{{ .GoPackage }}";

import "google/api/annotations.proto";
import "protoc-gen-openapiv2/options/annotations.proto";

// {{ .ServiceName }} - Auto-generated from Go structures
service {{ .ServiceName }} {
{{- range .Methods }}
    // {{ .Name }} RPC method
    rpc {{ .Name }}({{ .RequestType }}) returns ({{ .ResponseType }}) {
        {{- if .HTTPAnnotation }}
        option (google.api.http) = {
            {{ .HTTPAnnotation }}
        };
        {{- end }}
    }
{{- end }}
}

{{ range .Messages }}
// {{ .Name }} message
message {{ .Name }} {
{{- range .Fields }}
    {{- if .IsRepeated }}
    repeated {{ .Type }} {{ .Name }} = {{ .Number }};
    {{- else if .IsMap }}
    map<{{ .MapKeyType }}, {{ .MapValueType }}> {{ .Name }} = {{ .Number }};
    {{- else }}
    {{ .Type }} {{ .Name }} = {{ .Number }};
    {{- end }}
{{- end }}
}
{{ end }}`

type ProtoFile struct {
	Package      string
	GoPackage    string
	ServiceName  string
	Methods      []Method
	Messages     []Message
}

type Method struct {
	Name           string
	RequestType    string
	ResponseType   string
	HTTPAnnotation string
}

type Message struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name         string
	Type         string
	Number       int
	IsRepeated   bool
	IsMap        bool
	MapKeyType   string
	MapValueType string
}

func main() {
	// Get service definition from models
	def := models.GetServiceDefinition()

	// Create proto file structure
	protoFile := ProtoFile{
		Package:     def.Service.Package,
		GoPackage:   def.Service.GoPackage,
		ServiceName: def.Service.Name,
		Methods:     []Method{},
		Messages:    []Message{},
	}

	// Process methods
	for _, method := range def.Methods {
		m := Method{
			Name:         method.Name,
			RequestType:  getTypeName(method.Request),
			ResponseType: getTypeName(method.Response),
		}

		// Add HTTP annotation
		if method.HTTPMethod == "GET" {
			m.HTTPAnnotation = fmt.Sprintf(`get: "%s"`, method.HTTPPath)
		} else if method.HTTPMethod == "POST" {
			m.HTTPAnnotation = fmt.Sprintf(`post: "%s"
            body: "*"`, method.HTTPPath)
		}

		protoFile.Methods = append(protoFile.Methods, m)

		// Process request and response messages
		protoFile.Messages = append(protoFile.Messages, generateMessage(method.Request))
		protoFile.Messages = append(protoFile.Messages, generateMessage(method.Response))
	}

	// Add common messages
	protoFile.Messages = append(protoFile.Messages, generateMessage(models.ProcessingStats{}))
	protoFile.Messages = append(protoFile.Messages, generateMessage(models.ValidationError{}))

	// Remove duplicates
	protoFile.Messages = removeDuplicateMessages(protoFile.Messages)

	// Generate proto file
	outputPath := filepath.Join("src", "proto", "data_processor.proto")
	if err := generateProtoFile(protoFile, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating proto file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Proto file generated: %s\n", outputPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'buf generate' or 'make proto' to generate Go code")
	fmt.Println("2. Implement the service handlers")
}

func getTypeName(v interface{}) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func generateMessage(v interface{}) Message {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	msg := Message{
		Name:   t.Name(),
		Fields: []Field{},
	}

	// Skip empty structs
	if t.NumField() == 0 {
		return msg
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		protoTag := field.Tag.Get("proto")
		if protoTag == "" {
			continue
		}

		parts := strings.Split(protoTag, ",")
		number := 0
		fmt.Sscanf(parts[0], "%d", &number)

		f := Field{
			Name:   toSnakeCase(field.Name),
			Type:   goTypeToProto(field.Type),
			Number: number,
		}

		// Check for repeated fields
		if len(parts) > 1 && parts[1] == "repeated" {
			f.IsRepeated = true
			f.Type = goTypeToProto(field.Type.Elem())
		}

		// Check for map fields
		if field.Type.Kind() == reflect.Map {
			f.IsMap = true
			f.MapKeyType = goTypeToProto(field.Type.Key())
			f.MapValueType = goTypeToProto(field.Type.Elem())
		}

		msg.Fields = append(msg.Fields, f)
	}

	return msg
}

func goTypeToProto(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "bool"
	case reflect.Int32:
		return "int32"
	case reflect.Int64:
		return "int64"
	case reflect.Int:
		return "int32"
	case reflect.Float32:
		return "float"
	case reflect.Float64:
		return "double"
	case reflect.Slice:
		return goTypeToProto(t.Elem())
	case reflect.Map:
		return "map"
	case reflect.Struct:
		return t.Name()
	default:
		return "string"
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func removeDuplicateMessages(messages []Message) []Message {
	seen := make(map[string]bool)
	result := []Message{}

	for _, msg := range messages {
		if !seen[msg.Name] && msg.Name != "" {
			seen[msg.Name] = true
			result = append(result, msg)
		}
	}

	return result
}

func generateProtoFile(proto ProtoFile, outputPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add header comment
	fmt.Fprintln(file, "// Code generated by protogen. DO NOT EDIT.")
	fmt.Fprintln(file, "// source: src/internal/models/proto_models.go")
	fmt.Fprintln(file)

	// Parse and execute template
	tmpl, err := template.New("proto").Parse(protoTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(file, proto)
}