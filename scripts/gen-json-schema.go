package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Schema represents a JSON Schema
type Schema struct {
	Schema               string                 `json:"$schema,omitempty"`
	ID                   string                 `json:"$id,omitempty"`
	Type                 string                 `json:"type,omitempty"`
	Title                string                 `json:"title,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Items                interface{}            `json:"items,omitempty"`
	Properties           map[string]*Schema     `json:"properties,omitempty"`
	PatternProperties    map[string]*Schema     `json:"patternProperties,omitempty"`
	Required             []string               `json:"required,omitempty"`
	AdditionalProperties interface{}            `json:"additionalProperties,omitempty"`
	Enum                 []interface{}          `json:"enum,omitempty"`
	OneOf                []*Schema              `json:"oneOf,omitempty"`
	AnyOf                []*Schema              `json:"anyOf,omitempty"`
	Ref                  string                 `json:"$ref,omitempty"`
	Definitions          map[string]*Schema     `json:"definitions,omitempty"`
	MinItems             *int                   `json:"minItems,omitempty"`
	MaxItems             *int                   `json:"maxItems,omitempty"`
	Default              interface{}            `json:"default,omitempty"`
	Examples             []interface{}          `json:"examples,omitempty"`
	UniqueItems          *bool                  `json:"uniqueItems,omitempty"`
}

func main() {
	schema := generateRslintConfigSchema()
	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating schema: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("rslint-schema.json", output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schema file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated rslint-schema.json successfully")
}

func generateRslintConfigSchema() *Schema {
	return &Schema{
		Schema:      "http://json-schema.org/draft-07/schema#",
		ID:          "rslint-schema.json",
		Type:        "array",
		Title:       "Rslint Configuration",
		Description: "Configuration file for Rslint - a high-performance TypeScript/JavaScript linter",
		MinItems:    func() *int { i := 1; return &i }(),
		Items:       &Schema{Ref: "#/definitions/ConfigEntry"},
		Definitions: generateDefinitions(),
		Examples: []interface{}{
			[]interface{}{
				map[string]interface{}{
					"language": "javascript",
					"files":    []string{},
					"ignores":  []string{"node_modules/**", "dist/**", "*.test.ts"},
					"languageOptions": map[string]interface{}{
						"parserOptions": map[string]interface{}{
							"projectService": false,
							"project":        []string{"./tsconfig.json"},
						},
					},
					"rules": map[string]interface{}{
						"@typescript-eslint/no-unsafe-member-access": "error",
						"@typescript-eslint/no-floating-promises":     "warn",
						"@typescript-eslint/promise-function-async": []interface{}{
							"warn",
							map[string]interface{}{"allowAny": true},
						},
					},
					"plugins": []string{"@typescript-eslint"},
				},
			},
		},
	}
}

func generateDefinitions() map[string]*Schema {
	return map[string]*Schema{
		"ConfigEntry": {
			Type:        "object",
			Description: "A single configuration entry in the rslint.json array",
			Properties:  generateConfigEntryProperties(),
			Required:    []string{"language"},
			AdditionalProperties: false,
		},
		"LanguageOptions": {
			Type:        "object",
			Description: "Language-specific configuration options",
			Properties: map[string]*Schema{
				"parserOptions": {
					Ref:         "#/definitions/ParserOptions",
					Description: "Parser-specific configuration options",
				},
			},
			AdditionalProperties: false,
		},
		"ParserOptions": {
			Type:        "object",
			Description: "Parser-specific configuration options",
			Properties: map[string]*Schema{
				"projectService": {
					Type:        "boolean",
					Description: "Enable project service for typed linting (runs TypeScript language service behind the scene)",
					Default:     false,
				},
				"project": {
					Type:        "array",
					Description: "TypeScript project configuration files to use for typed linting",
					Items: &Schema{
						Type:        "string",
						Description: "Path to tsconfig.json file",
					},
					Examples: []interface{}{
						[]string{"./tsconfig.json"},
						[]string{"./packages/*/tsconfig.json", "./tsconfig.base.json"},
					},
				},
			},
			AdditionalProperties: false,
		},
		"Rules": {
			Type:        "object",
			Description: "Configuration for linting rules",
			PatternProperties: map[string]*Schema{
				"^@typescript-eslint/": {
					Ref: "#/definitions/RuleValue",
				},
				"^[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$": {
					Ref: "#/definitions/RuleValue",
				},
				"^[a-zA-Z0-9_-]+$": {
					Ref: "#/definitions/RuleValue",
				},
			},
			Properties: generateRuleProperties(),
			AdditionalProperties: &Schema{
				Ref: "#/definitions/RuleValue",
			},
		},
		"RuleValue": {
			Description: "Rule configuration value",
			OneOf: []*Schema{
				{
					Type:        "string",
					Enum:        []interface{}{"off", "warn", "error"},
					Description: "Simple rule severity level",
				},
				{
					Type:        "array",
					Description: "Array format rule configuration [severity, options]",
					MinItems:    func() *int { i := 1; return &i }(),
					MaxItems:    func() *int { i := 2; return &i }(),
					Items: []interface{}{
						map[string]interface{}{
							"type":        "string",
							"enum":        []interface{}{"off", "warn", "error"},
							"description": "Rule severity level",
						},
						map[string]interface{}{
							"type":                 "object",
							"description":          "Rule-specific options",
							"additionalProperties": true,
						},
					},
				},
				{
					Type:        "object",
					Description: "Object format rule configuration",
					Properties: map[string]*Schema{
						"level": {
							Type:        "string",
							Enum:        []interface{}{"off", "warn", "error"},
							Description: "Rule severity level",
						},
						"options": {
							Type:                 "object",
							Description:          "Rule-specific options",
							AdditionalProperties: true,
						},
					},
					AdditionalProperties: false,
				},
			},
		},
	}
}

func generateConfigEntryProperties() map[string]*Schema {
	return map[string]*Schema{
		"language": {
			Type:        "string",
			Description: "Programming language for this configuration entry",
			Enum:        []interface{}{"javascript"},
			Default:     "javascript",
		},
		"files": {
			Type:        "array",
			Description: "Additional files to include beyond those specified in tsconfig.json",
			Items: &Schema{
				Type:        "string",
				Description: "File pattern (glob) to include",
			},
			Default: []string{},
		},
		"ignores": {
			Type:        "array",
			Description: "File patterns to ignore when linting",
			Items: &Schema{
				Type:        "string",
				Description: "File pattern (glob) to ignore",
			},
			Examples: []interface{}{
				[]string{"node_modules/**", "dist/**", "*.test.ts"},
			},
		},
		"languageOptions": {
			Ref:         "#/definitions/LanguageOptions",
			Description: "Language-specific configuration options",
		},
		"rules": {
			Ref:         "#/definitions/Rules",
			Description: "Linting rules configuration",
			Default:     map[string]interface{}{},
		},
		"plugins": {
			Type:        "array",
			Description: "List of plugin names to enable",
			Items: &Schema{
				Type: "string",
				Enum: []interface{}{"@typescript-eslint"},
			},
			UniqueItems: func() *bool { b := true; return &b }(),
			Examples: []interface{}{
				[]string{"@typescript-eslint"},
			},
		},
	}
}

func generateRuleProperties() map[string]*Schema {
	rules := []string{
		"@typescript-eslint/adjacent-overload-signatures",
		"@typescript-eslint/array-type",
		"@typescript-eslint/await-thenable",
		"@typescript-eslint/class-literal-property-style",
		"@typescript-eslint/no-array-delete",
		"@typescript-eslint/no-base-to-string",
		"@typescript-eslint/no-confusing-void-expression",
		"@typescript-eslint/no-duplicate-type-constituents",
		"@typescript-eslint/no-floating-promises",
		"@typescript-eslint/no-for-in-array",
		"@typescript-eslint/no-implied-eval",
		"@typescript-eslint/no-meaningless-void-operator",
		"@typescript-eslint/no-misused-promises",
		"@typescript-eslint/no-misused-spread",
		"@typescript-eslint/no-mixed-enums",
		"@typescript-eslint/no-redundant-type-constituents",
		"@typescript-eslint/no-unnecessary-boolean-literal-compare",
		"@typescript-eslint/no-unnecessary-template-expression",
		"@typescript-eslint/no-unnecessary-type-arguments",
		"@typescript-eslint/no-unnecessary-type-assertion",
		"@typescript-eslint/no-unsafe-argument",
		"@typescript-eslint/no-unsafe-assignment",
		"@typescript-eslint/no-unsafe-call",
		"@typescript-eslint/no-unsafe-enum-comparison",
		"@typescript-eslint/no-unsafe-member-access",
		"@typescript-eslint/no-unsafe-return",
		"@typescript-eslint/no-unsafe-type-assertion",
		"@typescript-eslint/no-unsafe-unary-minus",
		"@typescript-eslint/no-unused-vars",
		"@typescript-eslint/no-useless-empty-export",
		"@typescript-eslint/no-var-requires",
		"@typescript-eslint/non-nullable-type-assertion-style",
		"@typescript-eslint/only-throw-error",
		"@typescript-eslint/prefer-as-const",
		"@typescript-eslint/prefer-promise-reject-errors",
		"@typescript-eslint/prefer-reduce-type-parameter",
		"@typescript-eslint/prefer-return-this-type",
		"@typescript-eslint/promise-function-async",
		"@typescript-eslint/related-getter-setter-pairs",
		"@typescript-eslint/require-array-sort-compare",
		"@typescript-eslint/require-await",
		"@typescript-eslint/restrict-plus-operands",
		"@typescript-eslint/restrict-template-expressions",
		"@typescript-eslint/return-await",
		"@typescript-eslint/switch-exhaustiveness-check",
		"@typescript-eslint/unbound-method",
		"@typescript-eslint/use-unknown-in-catch-callback-variable",
	}

	properties := make(map[string]*Schema)
	for _, rule := range rules {
		properties[rule] = &Schema{
			Ref: "#/definitions/RuleValue",
		}
	}
	return properties
}
