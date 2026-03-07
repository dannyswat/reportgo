package reportgo

import (
	"bytes"
	"testing"
	"text/template"
)

func TestGenerateToWriterMergesInlineData(t *testing.T) {
	engine := New()
	if err := engine.LoadTemplateFromString(`<?xml version="1.0" encoding="UTF-8"?>
<report xmlns="http://reportgo.io/schema/v1" version="1.0">
    <document orientation="portrait" format="A4"/>
    <sections>
        <section name="main">
            <text>{{.Title}}</text>
        </section>
    </sections>
</report>`); err != nil {
		t.Fatalf("LoadTemplateFromString returned error: %v", err)
	}

	var output bytes.Buffer
	if err := engine.GenerateToWriter(&output, map[string]interface{}{"Title": "Inline"}); err != nil {
		t.Fatalf("GenerateToWriter returned error: %v", err)
	}
	if engine.data["Title"] != "Inline" {
		t.Fatalf("expected inline data to be merged into engine state")
	}
	if output.Len() == 0 {
		t.Fatalf("expected PDF bytes to be written")
	}
}

func TestWithFuncMapOptionRegistersFunctions(t *testing.T) {
	engine := New(WithFuncMap(template.FuncMap{
		"shout": func(input string) string { return input + "!" },
	}))
	if err := engine.LoadTemplateFromString(`<?xml version="1.0" encoding="UTF-8"?>
<report xmlns="http://reportgo.io/schema/v1" version="1.0">
    <document orientation="portrait" format="A4"/>
    <sections>
        <section name="main">
            <text>{{shout .Word}}</text>
        </section>
    </sections>
</report>`); err != nil {
		t.Fatalf("LoadTemplateFromString returned error: %v", err)
	}

	var output bytes.Buffer
	if err := engine.GenerateToWriter(&output, map[string]interface{}{"Word": "go"}); err != nil {
		t.Fatalf("GenerateToWriter returned error with custom func map: %v", err)
	}
	if output.Len() == 0 {
		t.Fatalf("expected PDF bytes to be written")
	}
}
