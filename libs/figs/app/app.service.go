package app

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/thescaffold/gox-apps/libs/figs/app/file"
	"github.com/thescaffold/gox-apps/libs/figs/pkg/converter"
	"github.com/thescaffold/gox-apps/libs/figs/pkg/mapper"
	"github.com/thescaffold/gox-apps/libs/figs/pkg/store"
	"github.com/thescaffold/gox-apps/libs/figs/pkg/validator"
	"gopkg.in/yaml.v3"
)

// schemaHTTPClient fetches remote JSON-Schema documents referenced by
// meta.schemaUrl, mirroring TS JSONService which loads the schema via
// mediaService.loadAndGet(schemaUrl).
var schemaHTTPClient = &http.Client{Timeout: 10 * time.Second}

// fetchRemoteSchema GETs schemaUrl and parses it into a schema object (JSON, or
// YAML when the yaml validator is selected). Returns nil on any failure so the
// caller falls back to an inline meta["schema"] / no-op validation.
func fetchRemoteSchema(schemaURL string, validatorType validator.ProviderType) any {
	req, err := http.NewRequest(http.MethodGet, schemaURL, nil)
	if err != nil {
		return nil
	}
	resp, err := schemaHTTPClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var schema any
	if validatorType == validator.YAML {
		if err := yaml.Unmarshal(body, &schema); err != nil {
			return nil
		}
	} else if err := json.Unmarshal(body, &schema); err != nil {
		return nil
	}
	return schema
}

type AppService struct {
	fileService      *file.FileService  `inject:""`
	converterService *converter.Service `inject:""`
	mapperService    *mapper.Service    `inject:""`
	storeService     *store.Service     `inject:""`
	validatorService *validator.Service `inject:""`
}

func (s *AppService) GetHello() string {
	return "Hello World!"
}

func (s *AppService) FindFileByName(name string) (*file.File, error) {
	return s.fileService.FindByName(name)
}

// Init finds or creates a File record for the given payload (used by subscription handler).
func (s *AppService) Init(payload *Payload) (*file.File, error) {
	name, _ := payload.Meta["name"].(string)

	existing, err := s.fileService.FindByName(name)
	if err == nil && existing != nil {
		return existing, nil
	}

	inputJSON, _ := json.Marshal(payload.Input)
	outputJSON, _ := json.Marshal(payload.Output)
	metaJSON, _ := json.Marshal(payload.Meta)

	f := &file.File{
		Name:   name,
		Input:  inputJSON,
		Output: outputJSON,
		Meta:   metaJSON,
	}
	if err := s.fileService.Save(f); err != nil {
		return nil, err
	}
	return f, nil
}

// Process runs validate→map→convert→store on a File record (used by subscription handler).
func (s *AppService) Process(f *file.File) ([]any, *converter.Response, *store.Response, error) {
	var inputMap, outputMap, metaMap map[string]any
	_ = json.Unmarshal(f.Input, &inputMap)
	_ = json.Unmarshal(f.Output, &outputMap)
	_ = json.Unmarshal(f.Meta, &metaMap)

	payload := &Payload{Meta: metaMap, Input: inputMap, Output: outputMap}
	return s.runPipeline(payload)
}

// SaveFile runs the full pipeline for a POST /:name request.
func (s *AppService) SaveFile(name string, payload *Payload) (string, error) {
	payload.Meta["name"] = name
	if payload.Meta["bucket"] == nil {
		payload.Meta["bucket"] = "common"
	}

	_, converterResult, storeResult, err := s.runPipeline(payload)
	if err != nil {
		return "", err
	}
	if storeResult == nil || storeResult.URL == "" {
		return "", nil
	}

	outputMap := payload.Output
	if outputMap == nil {
		outputMap = map[string]any{}
	}
	outputMap["mime"] = converterResult.Mime
	outputMap["encoding"] = converterResult.Encoding
	outputMap["extension"] = converterResult.Extension

	inputJSON, _ := json.Marshal(payload.Input)
	outputJSON, _ := json.Marshal(outputMap)
	metaJSON, _ := json.Marshal(payload.Meta)

	f := &file.File{
		Name:   name,
		Input:  inputJSON,
		Output: outputJSON,
		Meta:   metaJSON,
		Url:    &storeResult.URL,
		Raw:    &storeResult.Raw,
	}
	if err := s.fileService.Save(f); err != nil {
		return "", err
	}
	return storeResult.URL, nil
}

// UpdateFile persists converter/store results back onto a File record (used by subscription handler).
func (s *AppService) UpdateFile(f *file.File, conv *converter.Response, st *store.Response) error {
	var outputMap map[string]any
	_ = json.Unmarshal(f.Output, &outputMap)
	outputMap["mime"] = conv.Mime
	outputMap["encoding"] = conv.Encoding
	outputMap["extension"] = conv.Extension
	outputJSON, _ := json.Marshal(outputMap)
	f.Output = outputJSON
	f.Url = &st.URL
	f.Raw = &st.Raw
	_, err := s.fileService.Update(f)
	return err
}

func (s *AppService) runPipeline(payload *Payload) ([]any, *converter.Response, *store.Response, error) {
	metaMap := payload.Meta
	inputMap := payload.Input
	outputMap := payload.Output

	validatorType := validator.JSON
	if v, ok := metaMap["validator"].(string); ok && v != "" {
		validatorType = validator.ProviderType(v)
	}
	mapperType := mapper.Data
	if v, ok := inputMap["type"].(string); ok && v != "" {
		mapperType = mapper.ProviderType(v)
	}
	converterType := converter.CSV
	if v, ok := outputMap["type"].(string); ok && v != "" {
		converterType = converter.ProviderType(v)
	}
	storeType := store.Local
	if v, ok := metaMap["store"].(string); ok && v != "" {
		storeType = store.ProviderType(v)
	}

	if schemaUrl, _ := metaMap["schemaUrl"].(string); schemaUrl != "" {
		// TS loads the schema remotely from meta.schemaUrl; fetch it and feed it
		// to the validator as meta["schema"] (unless an inline schema is present).
		if _, hasInline := metaMap["schema"]; !hasInline {
			if schema := fetchRemoteSchema(schemaUrl, validatorType); schema != nil {
				metaMap["schema"] = schema
			}
		}
		vPayload := &validator.Payload{Meta: metaMap, Input: inputMap, Output: outputMap}
		ok, err := s.validatorService.Use(validatorType).Validate(vPayload)
		if err != nil || !ok {
			return nil, nil, nil, err
		}
	}

	mPayload := &mapper.Payload{Meta: metaMap, Input: inputMap, Output: outputMap}
	mapResult, err := s.mapperService.Use(mapperType).Map(mPayload)
	if err != nil || len(mapResult) == 0 {
		return nil, nil, nil, err
	}

	converterResult, err := s.converterService.Use(converterType).Convert(mapResult...)
	if err != nil {
		return nil, nil, nil, err
	}

	sPayload := &store.Payload{Meta: metaMap, Input: inputMap, Output: outputMap}
	storeResult, err := s.storeService.Use(storeType).Store(sPayload, converterResult)
	if err != nil {
		return nil, nil, nil, err
	}

	return mapResult, converterResult, storeResult, nil
}
