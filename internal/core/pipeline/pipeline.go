package pipeline

type StepFunc func(*Context) error

type Pipeline struct {
	steps []StepFunc
	id    string
}

var Registry = make(map[string]*Pipeline)

type Builder struct {
	pipeline *Pipeline
}

func New(id string) *Builder {
	return &Builder{
		pipeline: &Pipeline{
			id:    id,
			steps: make([]StepFunc, 0), // 初始化空步骤列表
		},
	}
}

func (b *Builder) Step(f StepFunc) *Builder {
	b.pipeline.steps = append(b.pipeline.steps, f)
	return b
}

func (b *Builder) BuildAndRegister() *Pipeline {
	Registry[b.pipeline.id] = b.pipeline
	return b.pipeline
}

func (b *Builder) Build() *Pipeline {
	return b.pipeline
}

// Execute 执行pipeline中的所有步骤
func (p *Pipeline) Execute(ctx *Context) error {
	for _, step := range p.steps {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}
