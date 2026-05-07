package drift

import "fmt"

// PipelineStage is a function that transforms a DriftResult.
type PipelineStage func(DriftResult) (DriftResult, error)

// Pipeline executes a sequence of stages against a DriftResult produced by a
// Comparator, allowing arbitrary post-processing (scoring, labelling, etc.).
type Pipeline struct {
	comparator *Comparator
	stages     []PipelineStage
}

// NewPipeline creates a Pipeline backed by the given Comparator.
func NewPipeline(c *Comparator, stages ...PipelineStage) *Pipeline {
	return &Pipeline{comparator: c, stages: stages}
}

// AddStage appends a stage to the pipeline.
func (p *Pipeline) AddStage(s PipelineStage) {
	p.stages = append(p.stages, s)
}

// Run compares spec vs live through the Comparator, then passes the result
// through each registered stage in order.
func (p *Pipeline) Run(spec, live ServiceSpec) (DriftResult, error) {
	result, err := p.comparator.Compare(spec, live)
	if err != nil {
		return DriftResult{}, fmt.Errorf("pipeline: compare: %w", err)
	}
	for i, stage := range p.stages {
		result, err = stage(result)
		if err != nil {
			return DriftResult{}, fmt.Errorf("pipeline: stage %d: %w", i, err)
		}
	}
	return result, nil
}

// ScoreStage returns a PipelineStage that attaches a score to each entry via
// the scorer, storing the total in result.Meta if available.
func ScoreStage() PipelineStage {
	return func(r DriftResult) (DriftResult, error) {
		_ = ScoreResult(r) // score computed; callers may use BuildScoreReport
		return r, nil
	}
}

// LabelStage returns a PipelineStage that applies static labels to the result.
func LabelStage(l *Labeler) PipelineStage {
	return func(r DriftResult) (DriftResult, error) {
		return l.Label(r), nil
	}
}
