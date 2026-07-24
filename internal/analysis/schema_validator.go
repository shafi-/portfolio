package analysis

import (
	"encoding/json"

	"github.com/xeipuuv/gojsonschema"
)

// SchemaValidator validates analysis payloads against JSON schema
type SchemaValidator struct {
	schema *gojsonschema.Schema
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() (*SchemaValidator, error) {
	schemaLoader := gojsonschema.NewStringLoader(analysisSchema)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, WrapError(ErrCodeSchemaValidation, "failed to compile schema", err)
	}

	return &SchemaValidator{
		schema: schema,
	}, nil
}

// Validate validates an analysis input against the JSON schema
func (v *SchemaValidator) Validate(input AnalysisInput) error {
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return WrapError(ErrCodeSchemaValidation, "failed to marshal input", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(inputJSON)
	result, err := v.schema.Validate(documentLoader)
	if err != nil {
		return WrapError(ErrCodeSchemaValidation, "schema validation error", err)
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.Field()+": "+desc.Description())
		}
		return NewError(ErrCodeSchemaValidation, "validation failed: "+errors[0], nil)
	}

	return nil
}

// analysisSchema defines the JSON schema for analysis validation
const analysisSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["summary", "purpose", "architecture", "analyzed_at", "analyzed_git_head", "analyzer"],
  "properties": {
    "summary": { 
      "type": "string",
      "minLength": 1
    },
    "purpose": { 
      "type": "string",
      "minLength": 1
    },
    "architecture": { 
      "type": "string",
      "minLength": 1
    },
    "maturity": { 
      "type": ["string", "null"]
    },
    "strengths": { 
      "type": ["array", "null"], 
      "items": { "type": "string" } 
    },
    "weaknesses": { 
      "type": ["array", "null"], 
      "items": { "type": "string" } 
    },
    "reusable_components": { 
      "type": ["array", "null"], 
      "items": { "type": "string" } 
    },
    "notes": { 
      "type": ["string", "null"] 
    },
    "analyzed_at": { 
      "type": "string", 
      "format": "date-time" 
    },
    "analyzed_git_head": { 
      "type": "string",
      "minLength": 1
    },
    "analyzer": { 
      "type": "string",
      "minLength": 1
    },
    "features": {
      "type": ["array", "null"],
      "items": {
        "type": "object",
        "required": ["name"],
        "properties": {
          "name": { 
            "type": "string",
            "minLength": 1
          },
          "description": { 
            "type": ["string", "null"] 
          },
          "confidence": { 
            "type": ["number", "null"], 
            "minimum": 0, 
            "maximum": 1 
          }
        }
      }
    }
  },
  "additionalProperties": true
}`
