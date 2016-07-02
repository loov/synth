package realaudio

type Source interface {
	// Init is called once before starting rendering
	Init(format Format) error
	// Render is called for each buffer needing rendering.
	// The data is interleaved channels.
	// To stop rendering return ErrDone.
	// The size of the buffer is not a constant.
	Render(buffer []float32) error
}

type Sources []Source

func (sources Sources) Init(format Format) error {
	for _, source := range sources {
		if err := source.Init(format); err != nil {
			return err
		}
	}
	return nil
}

func (sources Sources) Render(buffer []float32) error {
	for _, source := range sources {
		if err := source.Render(buffer); err != nil {
			return err
		}
	}
	return nil
}

type RenderFunc func(buffer []float32) error

func (_ RenderFunc) Init(format Format) error           { return nil }
func (render RenderFunc) Render(buffer []float32) error { return render(buffer) }
