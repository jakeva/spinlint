package rules

// All is the ordered list of rules run during validation.
var All = []Rule{
	RequiredStageFields{},
	BrokenRequisiteRefs{},
	DuplicateRefIDs{},
	CircularDependencies{},
}
