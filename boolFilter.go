package govaluate

// boolFilter is used gather the information about evaluationStage`s which resulted boolean value
type boolFilter struct {
	boundTrue  *boundBoolFilter
	boundFalse *boundBoolFilter
}

func newBoolFilter(cap int) *boolFilter {
	return &boolFilter{
		boundTrue:  newBoundBoolFilter(cap, true),
		boundFalse: newBoundBoolFilter(cap, false),
	}
}

func (bf *boolFilter) getVarsCausing(value bool) map[string]bool {
	if value {
		return bf.boundTrue.vars
	} else {
		return bf.boundFalse.vars
	}
}

func (bf *boolFilter) pushStageResult(stage *evaluationStage) {
	// route boolean expressions to their filters
	if stage.result == true {
		bf.boundTrue.pushStageResult(stage)
		return
	}
	if stage.result == false {
		bf.boundFalse.pushStageResult(stage)
		return
	}

	// ternary expression is split to 2 actual expressions: TERNARY_TRUE and TERNARY_FALSE
	// so if expression originally was: `a > b ? c : d` then there are 2 expressions in result:
	// TERNARY_TRUE(x1, c) and TERNARY_FALSE(x2, d), and c is `true` part while d is `false` part
	if stage.symbol == TERNARY_TRUE && stage.rightStage != nil {
		bf.boundTrue.pushStageResult(stage.rightStage)
		return
	}
	if stage.symbol == TERNARY_FALSE && stage.rightStage != nil {
		bf.boundFalse.pushStageResult(stage.rightStage)
		return
	}
}

// postprocessStages gathers information about names of variables which cause true and false
// it is called once after all evaluationStage`s are processed
func (bf *boolFilter) postprocessStages() {
	bf.boundTrue.postprocessStages()
	bf.boundFalse.postprocessStages()
}

// boundBoolFilter is boolFilter part that is bound to specific boolean value: true or false
type boundBoolFilter struct {
	value   bool
	avoid   bool
	stages  []*evaluationStage
	vars    map[string]bool
	visited map[*evaluationStage]bool
}

func newBoundBoolFilter(cap int, value bool) *boundBoolFilter {
	return &boundBoolFilter{
		value:   value,
		avoid:   !value,
		stages:  make([]*evaluationStage, 0, cap),
		vars:    make(map[string]bool, cap),
		visited: make(map[*evaluationStage]bool, cap),
	}
}

func (bbf *boundBoolFilter) collectVarsInfo(stage *evaluationStage) {
	// avoid visiting stages repeatedly
	if bbf.visited[stage] {
		return
	}
	bbf.visited[stage] = true

	// don't go down to subtrees those results differ from bound value
	if stage.result == bbf.avoid {
		return
	}

	if stage.variableName != nil {
		bbf.vars[*stage.variableName] = true
	}
	if stage.leftStage != nil {
		bbf.collectVarsInfo(stage.leftStage)
	}
	if stage.rightStage != nil {
		bbf.collectVarsInfo(stage.rightStage)
	}
}

func (bbf *boundBoolFilter) postprocessStages() {
	for _, stage := range bbf.stages {
		bbf.collectVarsInfo(stage)
	}
}

func (bbf *boundBoolFilter) pushStageResult(stage *evaluationStage) {
	bbf.stages = append(bbf.stages, stage)
}
