package processor

import "encoding/json"

type ProcessorFunc[T any] func(data T) error

type ProcessorImpl[T any] interface {
	Parse(payload string) (*T, error)
	Process(payload string, fn ProcessorFunc[T]) error
}

type Processor[T any] struct {
}

func NewProcessor[T any]() *Processor[T] {
	return &Processor[T]{}
}

func (p *Processor[T]) Parse(payload string) (*T, error) {
	// parse from json
	var data T
	err := json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return nil, err
	}
	return &data, nil

}

func (p *Processor[T]) Process(payload string, fn ProcessorFunc[T]) error {
	data, err := p.Parse(payload)
	if err != nil {
		return err
	}

	// do something with data
	return fn(*data)
}
