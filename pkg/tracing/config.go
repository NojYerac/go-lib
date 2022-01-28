package tracing

type Configuration struct {
	ExporterType string  `config:"trace_exporter_type" validate:"oneof=jaeger stdout file noop"`
	SampleRatio  float64 `config:"trace_sample_ratio" validate:"required_if=ExporterType jaeger,max=1,min=0"`
	FilePath     string  `config:"trace_file_path" validate:"omitempty,required_if=ExporterType file,file"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		ExporterType: "stdout",
		SampleRatio:  0.5,
	}
}
