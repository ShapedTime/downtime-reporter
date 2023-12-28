package core

type Transformer interface {
	Transform() Result
}

func (r Result) Transform() Result {
	return r
}

type TransformerSelector struct {
	transformers map[string]Transformer
}

func (s TransformerSelector) Register(name string, transformer Transformer) {
	s.transformers[name] = transformer
}
