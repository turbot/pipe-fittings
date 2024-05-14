package constants

// ModUpdateStrategy controls how mods are updated. It is one of:
// 1. full - check everything for both latest and accuracy
// 2. latest - update everything to latest, but only branches - not tags - are commit checked (which is the same as latest)
// 3. development - update branches and broken constraints to latest, leave satisfied constraints unchanged
// 4. minimal - only update broken constraints, do not check branches for new commits
type ModUpdateStrategy int

const (
	ModUpdateFull        = "full"
	ModUpdateLatest      = "latest"
	ModUpdateDevelopment = "development"
	ModUpdateMinimal     = "minimal"
)

const (
	ModUpdateIdFull ModUpdateStrategy = iota // default for command
	ModUpdateIdLatest
	ModUpdateIdDevelopment
	ModUpdateIdMinimal
)

var ModUpdateStrategyIds = map[ModUpdateStrategy][]string{
	ModUpdateIdFull:        {ModUpdateFull},
	ModUpdateIdLatest:      {ModUpdateLatest},
	ModUpdateIdDevelopment: {ModUpdateDevelopment},
	ModUpdateIdMinimal:     {ModUpdateMinimal},
}

func FlagValues[T comparable](mappings map[T][]string) []string {
	var res = make([]string, 0, len(mappings))
	for _, v := range mappings {
		res = append(res, v[0])
	}
	return res

}
